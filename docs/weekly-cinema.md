# Weekly cinema segment

Last updated: 2026-10-18. Status: reviewed profile, evidence bundles and evaluation runner ready; live scoring blocked on API credentials. No production cinema implementation yet.

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

Originally documented as a proposal; the user subsequently requested running the taste-profile experiment. The first pass below has now been prepared locally.

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
4. Review the draft profile, then run the held-out ranking comparison before wiring the weekly segment into production.

## First taste-profile pass (2026-10-18)

- Reproducible stdlib-only preparation: `python3 scripts/prepare_cinema_experiment.py`.
- Artifacts live in `data/cinema/`: full normalized film ratings, separate normalized bookmarks, training ratings, label-free held-out candidates, separate held-out labels, and a source-hash/split manifest. Source CSVs are unchanged; missing scalar metadata remains null and missing lists remain empty.
- Split: 483 films into 386 training / 97 held out, using a fixed hash ordering within each rating stratum and approximately 20% held out. All 63 film bookmarks remain separate from training evidence. TV and other title types are excluded from this first pass.
- Codex drafted eight unreviewed hypotheses in `data/cinema/taste_profile.draft.json`, with a readable, rated-example version in `data/cinema/PROFILE_REVIEW.md`. Only training rows were inspected for profile construction. No external model API calls or enrichment were performed; no model API credentials were configured in the shell.
- Initial finding: useful film-level contrasts exist within genres and directors. Subsequent user review confirmed science fiction and eccentric/absurdist comedy as positive preferences; film-level exceptions still matter.
- Validation checked split completeness/disjointness, exclusion of labels and rating dates from held-out candidates, multi-director parsing, and that every profile evidence reference belongs to training.
- This is a draft-generation result, not evidence of predictive accuracy. Next: user review, freeze profile, implement a shared training-only example selector, then compare examples-only versus profile-plus-examples and LLM versus Jev against a simple baseline. Candidate bundles and scoring calls have not yet been produced.
- Evaluation caveat: the title-level split permits franchise overlap. It does not establish generalization to unseen franchises or future taste. Keep the held-out labels uninspected until predictions are fixed.

## Profile review feedback (2026-10-18)

- Confirmed explicit preferences: science fiction and eccentric/absurdist comedy.
- Anime could be a useful indicator, but remains tentative. Animation overall is too broad. Removed non-anime examples from that profile entry; candidate anime classification will need reliable metadata because the exports do not explicitly tag it.
- Removed action and drama as broad positive taste signals. This does not indicate dislike, and individual ratings remain available for relevant-example retrieval.
- User reports relatively few anime screenings in Irish cinemas. Record this as user-provided availability context, separate from taste fit; no current listing coverage was verified.
- Further user review rejected the franchise and music hypotheses as uninformative personal selection signals. Removed both from the JSON and readable profile; retained individual ratings for example retrieval. Do not infer a music/musical or franchise bonus.
- Updated both local profile artifacts. With user authorization to proceed, moved the director observation into example-retrieval guidance. The pilot profile contains confirmed science fiction and eccentric/absurdist comedy plus tentative anime. Enjoyment versus admiration and rewatch interest remain unanswered; the experiment predicts personal ratings, not attendance.

## Offline runner and frozen pilot (2026-10-18)

Implemented `scripts/cinema_experiment.py` using Python's standard library. The first 12-candidate pilot is frozen locally at `data/cinema/runs/pilot-v1/inputs.json`, containing 24 request bundles and profile/source snapshots. Selection uses a fixed hash ordering without inspecting held-out labels. Build refuses to overwrite an existing run.

- Retrieval: shared director contributes 2, genre Jaccard similarity contributes 0–1. Select up to three ratings >=8 and three <=5 with positive metadata overlap; break ties by IMDb ID. No unrelated filler. Include the training rating distribution in both arms to disclose the deliberately balanced example selection.
- Both arms receive identical candidate metadata and rated examples. Only the profile arm receives explicit preferences and interpretation rules; the profile's additional supporting-film lists are excluded to avoid confounding preference text with more examples. Both providers receive the same evidence bundles, with provider-specific output instructions.
- Use supplied metadata only. Anime classification and absurdist tone are absent from the exports, so this first pass cannot fully test those preferences. Consider sourced enrichment as a separate subsequent experiment rather than silently relying on model film knowledge.
- LLM predicts a rating from 1–10 with a reason. Jev uses an ordered five-level taste-fit rubric, scoring 0–4. Compare ranking discrimination across providers; do not treat Jev scores as calibrated ratings.
- Evaluation reports high-versus-low AUC (>=8 versus <=5; ties count half) and rating MAE where applicable. Baselines are the training mean and aggregate IMDb rating. Per-film results support error inspection. Report sample sizes; 12 films are exploratory, not a definitive benchmark.
- Scoring resumes from saved predictions and checks input/model/endpoint identity. Evaluation requires complete predictions for every requested provider and both arms before opening labels. No real held-out labels have been inspected and no real predictive metrics are available yet.
- Five local tests passed, covering retrieval, label exclusion, frozen inputs, resumability, provider parsing, score validation, metrics and the complete-predictions evaluation boundary. Provider responses were mocked; live API compatibility remains unverified.
- Attempted the LLM scoring command; it stopped before network access because `DIGESTBOT_OPENAI_API_KEY` is unset. Neither LLM nor Jev credentials are configured. Next step: configure credentials in the execution environment and run the frozen pilot. Do not paste secrets into the conversation.

Run from the repository root:

```sh
# Already built; use a fresh run path for a new pilot.
python3 scripts/cinema_experiment.py build --run data/cinema/runs/pilot-v1 --limit 12

# Uses DIGESTBOT_OPENAI_API_KEY, optional DIGESTBOT_LLM_MODEL / DIGESTBOT_LLM_BASE_URL.
python3 scripts/cinema_experiment.py score --run data/cinema/runs/pilot-v1 --provider llm

# Uses DIGESTBOT_JEV_API_KEY. Optional --model and --endpoint for either provider.
python3 scripts/cinema_experiment.py score --run data/cinema/runs/pilot-v1 --provider jev

# Complete both providers before opening labels for a direct comparison.
python3 scripts/cinema_experiment.py evaluate --run data/cinema/runs/pilot-v1 --providers llm jev

# An LLM-only pilot can instead use --providers llm.
python3 -m unittest discover -s scripts -p 'test_cinema_experiment.py'
```

Data and run artifacts remain local under `data/` (ignored by the existing Git configuration); scripts and this workflow documentation are versionable. After inspecting pilot results, treat those 12 films as validation data for any tuning, not as an untouched test set.
