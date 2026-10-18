# Weekly cinema segment

Last updated: 2026-10-18. Status: requirements and proposed ranking experiment documented; no implementation yet.

## Goal

Keep the normal morning digest, adding a weekly segment that highlights interesting upcoming screenings at the user's 3–4 favourite cinemas. A small, cinema-specific solution is fine; no need for a generic cinema platform.

## Requirements

- Support weekly segments alongside daily ones. Start simply: a flag/config option and a hardcoded Monday are acceptable. Evaluate the weekday in the configured timezone (currently defaults to Europe/Dublin).
- Include the weekly segment in the same morning digest. The executable currently produces one digest per invocation; weekday gating can use the existing daily invocation without building a scheduler.
- Obtain listings for the chosen cinemas. Investigate APIs/feeds or scrape individual sites once the cinemas are identified; no source selected yet.
- Recommend screenings based on personal taste using `data/imdb/ratings.csv` and `data/imdb/watchlist.csv`. Watchlist membership means bookmarked, not a commitment to watch or a positive rating.
- One-off preprocessing of these exports into useful ranking input is welcome. Explore both LLM ranking and Jev scoring; Jev would be particularly interesting if it works well.
- Keep this document lightweight and update it at the end of each work session with decisions, findings, and the next step.

## Existing pieces and data

- Segment interface: `internal/segment/segment.go`; registration/selection: `cmd/digestbot/main.go`; configuration: `internal/config/`; composition: `internal/digest/`.
- Existing Jev news ranker: `internal/news/jev.go`. Uses TypeSafe System One with score questions and an ordered rubric. Reuse the integration pattern, not the news significance rubric.
- IMDb exports inspected: 604 ratings (483 Movie entries) and 92 bookmarks (63 Movie entries). Both also contain television and other title types.
- Available metadata includes IMDb ID (`Const`), title, year, genres, directors, runtime, and aggregate IMDb rating. Personal ratings and dates are available. No plot summaries in these exports.

## Ranking ideas to evaluate (not decisions)

- Offline: normalize by IMDb ID; derive a compact, reviewable taste profile grounded in positive and negative examples. Start with film ratings; decide whether television adds useful evidence. Enrich selected examples with film descriptions if needed.
- Weekly: resolve listing titles to films, enrich candidates, rank each unique film, then attach its cinema showtimes and booking links.
- Give the ranker a taste profile plus relevant liked/disliked examples for each candidate. Treat bookmarks separately as weak evidence of curiosity.
- Separate predicted enjoyment from reasons to see something in a cinema (e.g. a favourite being reissued). Decide how to handle already-rated films rather than automatically suppressing them.
- Compare LLM and Jev on the same candidates/evidence. Historical ratings can provide a held-out taste check, with held-out titles excluded from profile construction and example retrieval. Actual screening-interest judgments are needed too: enjoyment and willingness to go out differ.
- Allow no recommendations when nothing seems worthwhile; avoid filling a fixed quota with weak matches.

## Proposed experiment: taste profile + relevant examples

Discussed with the user; document only for now. No profile generation, enrichment, or model calls have been authorized as part of this session.

1. Normalize film ratings by IMDb ID into a separate local data file, retaining all ratings rather than compressing them away.
2. Have an LLM draft a short JSON taste profile (roughly 5–10 hypotheses) from titles, ratings, directors, and genres. Look for contrasts between similar films rated differently, supporting examples, counterexamples, and uncertainty. Start with existing metadata; try sourced synopses/tone/style descriptions if relying on model knowledge of films proves inadequate.
3. Review the hypotheses with the user. Keep inferred patterns separate from explicit user preferences; a rating alone does not explain why someone liked a film.
4. For each candidate, assemble the profile, candidate metadata, and a small selection of relevant positive and negative rated examples. Example selection is still to be designed.
5. Feed the same bundle to Jev as its `state` with a taste-fit scoring question and descriptive rubric, or to an LLM with ranking/explanation instructions. This is structured input, not model training.

Proposed input shape (placeholder content, not an actual inference about the user):

```json
{
  "taste_profile": {
    "preferences": [
      {
        "pattern": "A hypothesis about taste",
        "strength": "tentative",
        "supporting_films": ["<IMDb ID>"],
        "counterexamples": ["<IMDb ID>"],
        "uncertainty": "Possible alternative explanation"
      }
    ],
    "explicit_preferences": [],
    "interpretation_rules": [
      "Watchlist membership indicates curiosity, not endorsement",
      "An unrated film is not evidence of dislike"
    ]
  },
  "candidate": {
    "imdb_id": "<IMDb ID>",
    "title": "Candidate film",
    "director": "Director",
    "synopsis": "Sourced description if available",
    "genres": ["Drama"],
    "on_watchlist": true,
    "user_rating": null
  },
  "relevant_ratings": [
    {
      "imdb_id": "<IMDb ID>",
      "title": "Related rated film",
      "rating": 9,
      "description": "Description if available"
    }
  ]
}
```

Artifacts would be a readable/editable local JSON profile, separate normalized ratings, and candidate-specific request bundles. Paths and exact schema remain undecided. Preserve missing metadata rather than inventing it.

Evaluation: hold out ratings before constructing the profile or retrieving examples; also hide each held-out candidate's personal rating from model input. Compare profile-plus-examples against examples alone, and compare LLM with Jev using the same evidence. Check whether recommendations distinguish liked from disliked films, not merely whether the profile sounds convincing. Follow with user judgments on real listings to test cinema-going interest. Keep enjoyment and screening appeal separate; no rubric or weighting has been selected yet.

## Open questions / next session

1. Clarify whether ratings reflect enjoyment, admiration, or both, and how important rewatches are compared with discoveries. These questions remain unanswered; do not encode illustrative preferences as user facts.
2. Get the cinema names/URLs and inspect their listing sources and publication horizons.
3. Decide the coverage window (next seven days is a starting proposal), result count, and desired output details.
4. Build a small offline ranking experiment before wiring the weekly segment into production. No API calls with IMDb data or preprocessing artifacts have been made yet.
