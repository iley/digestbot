import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import cinema_experiment as exp


def film(id, rating=None, genres=None, directors=None):
    result = {"imdb_id": id, "title": id, "genres": genres or ["Sci-Fi"],
              "directors": directors or [], "imdb_rating": 7.0}
    if rating is not None:
        result["user_rating"] = rating
    return result


class ExperimentTest(unittest.TestCase):
    def test_retrieval_balances_related_examples_and_excludes_self(self):
        candidate = film("target", directors=["A"])
        training = [film("target", 10), film("high", 9), film("low", 2),
                    film("director", 8, ["Drama"], ["A"]),
                    film("unrelated", 1, ["Horror"])]
        found = exp.examples(candidate, training)
        self.assertEqual([r["imdb_id"] for r in found], ["director", "high", "low"])

    def test_metrics_ties_and_ordinal_scale(self):
        labels = {"a": 9, "b": 2, "c": 7}
        self.assertEqual(exp.metrics({"a": 8, "b": 3, "c": 7}, labels, True)["high_vs_low_auc"], 1)
        tied = exp.metrics({i: 2 for i in labels}, labels, False)
        self.assertEqual(tied["high_vs_low_auc"], 0.5)
        self.assertNotIn("mae", tied)
        self.assertIsNone(exp.metrics({"c": 3}, labels, False)["high_vs_low_auc"])

    def test_invalid_provider_scores(self):
        for value in (True, None, float("nan"), float("inf"), 11, 0):
            with self.assertRaises(ValueError):
                exp.valid_score(value)

    def test_provider_parsers(self):
        response = {"model": "mock", "choices": [{"message": {"content":
                    '```json\n{"rating": 8, "reason": "evidence"}\n```'}}]}
        with patch.object(exp, "post", return_value=response):
            self.assertEqual(exp.predict("llm", "mock", "https://example.invalid", "fake", {}, exp.INSTRUCTIONS)["score"], 8)
        with patch.object(exp, "post", return_value={"answers": {"taste_fit": {"type": "score", "score": 0}}}):
            self.assertEqual(exp.predict("jev", "mock", "https://example.invalid", "fake", {}, exp.INSTRUCTIONS)["score"], 0)

    def test_frozen_build_resume_and_evaluation_boundary(self):
        with tempfile.TemporaryDirectory() as temp:
            data = Path(temp)
            run = data / "run"
            exp.write(data / "training_ratings.json", [film("high", 9), film("low", 2)])
            # Deliberately contaminate candidate input: projection must strip labels.
            exp.write(data / "heldout_candidates.json", [dict(film("target", 1), date_rated="secret")])
            exp.write(data / "taste_profile.draft.json", {"preferences": [], "explicit_preferences": [], "interpretation_rules": []})
            exp.write(data / "manifest.json", {"test": True})
            # No labels file exists: build and score must not need it.
            exp.build(data, run, 1)
            frozen = exp.read(run / "inputs.json")
            first, second = frozen["bundles"]
            self.assertEqual(first["state"]["relevant_ratings"], second["state"]["relevant_ratings"])
            self.assertNotIn("taste_profile", first["state"])
            self.assertNotIn("user_rating", first["state"]["candidate"])
            self.assertNotIn("date_rated", first["state"]["candidate"])
            with self.assertRaises(FileExistsError):
                exp.build(data, run, 1)
            with patch.dict(exp.os.environ, {"DIGESTBOT_OPENAI_API_KEY": "fake"}), patch.object(exp, "predict", return_value={"score": 7, "reason": "mock"}) as call:
                exp.score(run, "llm", "mock", "https://example.invalid")
                exp.score(run, "llm", "mock", "https://example.invalid")
                self.assertEqual(call.call_count, 2)
                with self.assertRaises(ValueError):
                    exp.score(run, "llm", "changed", "https://example.invalid")
            result = exp.read(run / "predictions.llm.json")
            exp.write(run / "predictions.llm.json", dict(result, predictions=result["predictions"][:1]))
            with self.assertRaisesRegex(ValueError, "Complete both variants"):
                exp.evaluate(data, run, ["llm"])
            exp.write(run / "predictions.llm.json", result)
            exp.write(data / "heldout_labels.json", {"target": 8})
            exp.evaluate(data, run, ["llm"])
            self.assertEqual(exp.read(run / "evaluation.json")["metrics"]["llm.examples_only"]["mae"], 1)


if __name__ == "__main__":
    unittest.main()
