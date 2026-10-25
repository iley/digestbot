#!/usr/bin/env python3
"""Prepare a reproducible, local-only film taste experiment (stdlib only)."""
import csv
import hashlib
import json
from collections import Counter, defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "data/cinema"
SPLIT_SEED = "cinema-taste-v1"


def write(name, value):
    (OUT / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n")


def normalize(name):
    films = {}
    with (ROOT / f"data/imdb/{name}.csv").open(newline="", encoding="utf-8-sig") as source:
        for row in csv.DictReader(source):
            if row["Title Type"] != "Movie":
                continue
            film = {"imdb_id": row["Const"], "title": row["Title"],
                    "original_title": row["Original Title"] or None,
                    "year": int(row["Year"]) if row["Year"] else None,
                    "genres": row["Genres"].split(", ") if row["Genres"] else [],
                    "directors": [v.strip() for v in row["Directors"].split(",") if v.strip()],
                    "runtime_minutes": int(row["Runtime (mins)"]) if row["Runtime (mins)"] else None,
                    "imdb_rating": float(row["IMDb Rating"]) if row["IMDb Rating"] else None,
                    "url": row["URL"] or None}
            if name == "ratings":
                film.update(user_rating=int(row["Your Rating"]), date_rated=row["Date Rated"] or None)
            else:
                # Bookmarks and their free text are not training evidence.
                film.update(on_watchlist=True, created=row["Created"] or None)
            if film["imdb_id"] in films:
                raise ValueError(f"Duplicate IMDb ID in {name}: {film['imdb_id']}")
            films[film["imdb_id"]] = film
    return sorted(films.values(), key=lambda f: f["imdb_id"])


def main():
    OUT.mkdir(parents=True, exist_ok=True)
    ratings, bookmarks = normalize("ratings"), normalize("watchlist")
    strata = defaultdict(list)
    for film in ratings:
        strata[film["user_rating"]].append(film)
    held_out = set()
    for films in strata.values():
        ordered = sorted(films, key=lambda f: hashlib.sha256(
            f"{SPLIT_SEED}:{f['imdb_id']}".encode()).hexdigest())
        held_out.update(f["imdb_id"] for f in ordered[:round(len(films) * 0.2)])
    training = [f for f in ratings if f["imdb_id"] not in held_out]
    test = [f for f in ratings if f["imdb_id"] in held_out]
    # Keep labels in a separate artifact, never in candidate request metadata.
    candidates = [{k: v for k, v in f.items() if k not in ("user_rating", "date_rated")}
                  for f in test]
    write("ratings.normalized.json", ratings)
    write("watchlist.normalized.json", bookmarks)
    write("training_ratings.json", training)
    write("heldout_candidates.json", candidates)
    write("heldout_labels.json", {f["imdb_id"]: f["user_rating"] for f in test})
    manifest = {"schema_version": 1, "split_seed": SPLIT_SEED,
                "method": "Within each rating, sort SHA256(seed:IMDb ID); hold out round(n * 0.2).",
                "scope": "Title Type == Movie; no enrichment; bookmarks excluded from training",
                "source_sha256": {name: hashlib.sha256((ROOT / f"data/imdb/{name}.csv").read_bytes()).hexdigest()
                                  for name in ("ratings", "watchlist")},
                "ratings_count": len(ratings), "training_count": len(training),
                "heldout_count": len(test), "bookmark_count": len(bookmarks),
                "training_rating_histogram": dict(sorted(Counter(f["user_rating"] for f in training).items()))}
    write("manifest.json", manifest)
    print(json.dumps(manifest, indent=2))


if __name__ == "__main__":
    main()
