#!/usr/bin/env python3
"""
Seeds data/problems/ and data/submissions/ with example fixture data.
Run syncer afterwards to load this into Postgres.
The meta.json created within submissions/{uuid} is only for example fixture data that needs to be synced.
Future added submissions will be added without this.

Usage:
    python3 scripts/seed.py [--data-dir data] [--force]
"""

import argparse
import json
import shutil
import uuid
from pathlib import Path

PROBLEMS = [
    {
        "id": "sum-two-numbers",
        "time_limit_ms": 1000,
        "memory_limit_mb": 128,
        "problem_type": "standard",
        "statement": (
            "# Sum Two Numbers\n\n"
            "Given two integers $a$ and $b$ output their sum."
        ),
        "testcases": [
            ("3 4\n", "7\n"),
            ("-1 1\n", "0\n"),
            ("100 250\n", "350\n"),
        ],
    }
]

# (problem_id, lang, filename, source)
SUBMISSIONS = [
    # --- AC: correct C++ solution ---
    (
        "sum-two-numbers",
        "C++",
        "main.cpp",
        "#include <bits/stdc++.h>\n"
        "using namespace std;\n"
        "int main() {\n"
        "    int a, b;\n"
        "    cin >> a >> b;\n"
        "    cout << a + b << endl;\n"
        "    return 0;\n"
        "}\n",
    ),

    # --- CE: missing semicolon, should fail to compile ---
    (
        "sum-two-numbers",
        "C++",
        "main.cpp",
        "#include <bits/stdc++.h>\n"
        "using namespace std;\n"
        "int main() {\n"
        "    int a, b\n" # <- missing semicolon here
        "    cin >> a >> b;\n"
        "    cout << a + b << endl;\n"
        "    return 0;\n"
        "}\n",
    ),

    # --- AC: correct C solution ---
    (
        "sum-two-numbers",
        "C",
        "main.c",
        "#include <stdio.h>\n"
        "int main() {\n"
        "    int a, b;\n"
        "    scanf(\"%d %d\", &a, &b);\n"
        "    printf(\"%d\\n\", a + b);\n"
        "    return 0;\n"
        "}\n",
    ),

    # --- AC: correct Python solution ---
    (
        "sum-two-numbers",
        "Python",
        "solution.py",
        "a, b = map(int, input().split())\n"
        "print(a + b)\n",
    ),

    # --- AC: correct Java solution (must be Main.java / class Main) ---
    (
        "sum-two-numbers",
        "Java",
        "Main.java",
        "import java.util.Scanner;\n"
        "public class Main {\n"
        "    public static void main(String[] args) {\n"
        "        Scanner sc = new Scanner(System.in);\n"
        "        int a = sc.nextInt();\n"
        "        int b = sc.nextInt();\n"
        "        System.out.println(a + b);\n"
        "    }\n"
        "}\n",
    ),

    # --- WA: compiles and runs fine, but produces the wrong answer (subtracts instead of adds) ---
    (
        "sum-two-numbers",
        "C++",
        "main.cpp",
        "#include <bits/stdc++.h>\n"
        "using namespace std;\n"
        "int main() {\n"
        "    int a, b;\n"
        "    cin >> a >> b;\n"
        "    cout << a - b << endl;\n"
        "    return 0;\n"
        "}\n",
    ),

    # --- RTE: out-of-bounds vector access ---
    (
        "sum-two-numbers",
        "C++",
        "main.cpp",
        "#include <bits/stdc++.h>\n"
        "using namespace std;\n"
        "int main() {\n"
        "    vector<int> x;\n"
        "    cout << x[100];\n"
        "}\n",
    ),

    # --- TLE: infinite loop ---
    (
        "sum-two-numbers",
        "C++",
        "main.cpp",
        "#include <bits/stdc++.h>\n"
        "using namespace std;\n"
        "int main() {\n"
        "    while (1) {\n"
        "        // nothing\n"
        "    }\n"
        "}\n",
    ),

    # --- MLE: allocates a huge 50000x50000 int grid ---
    (
        "sum-two-numbers",
        "C++",
        "main.cpp",
        "#include <bits/stdc++.h>\n"
        "using namespace std;\n"
        "int main() {\n"
        "    vector<vector<int>> grid(50000, vector<int>(50000));\n"
        "    cout << grid[0][0];\n"
        "}\n",
    ),
]


def write_problem(problems_dir: Path, problem: dict) -> None:
    problem_dir = problems_dir / problem["id"]
    testcases_dir = problem_dir / "testcases"
    testcases_dir.mkdir(parents=True, exist_ok=True)

    meta = {
        "id": problem["id"],
        "time_limit_ms": problem["time_limit_ms"],
        "memory_limit_mb": problem["memory_limit_mb"],
        "problem_type": problem["problem_type"],
    }
    (problem_dir / "meta.json").write_text(json.dumps(meta, indent=2) + "\n")
    (problem_dir / "problem.md").write_text(problem["statement"])

    for i, (input_text, output_text) in enumerate(problem["testcases"], start=1):
        (testcases_dir / f"{i}.in").write_text(input_text)
        (testcases_dir / f"{i}.out").write_text(output_text)

    print(f"  wrote problem: {problem['id']} ({len(problem['testcases'])} testcases)")


def write_submission(submissions_dir: Path, problem_id: str, lang: str,
                      source_filename: str, source: str) -> None:
    submission_id = str(uuid.uuid4())
    submission_dir = submissions_dir / submission_id
    submission_dir.mkdir(parents=True, exist_ok=True)

    meta = {"problem_id": problem_id, "lang": lang}
    (submission_dir / "meta.json").write_text(json.dumps(meta, indent=2) + "\n")
    (submission_dir / source_filename).write_text(source)

    print(f"  wrote submission: {submission_id} -> {problem_id} ({lang})")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--data-dir", default="data",
                         help="root data directory (default: data)")
    parser.add_argument("--force", action="store_true",
                         help="wipe existing problems/submissions dirs first")
    args = parser.parse_args()

    data_dir = Path(args.data_dir)
    problems_dir = data_dir / "problems"
    submissions_dir = data_dir / "submissions"

    if args.force:
        shutil.rmtree(problems_dir, ignore_errors=True)
        shutil.rmtree(submissions_dir, ignore_errors=True)

    print(f"Seeding problems into {problems_dir}/")
    for problem in PROBLEMS:
        write_problem(problems_dir, problem)

    print(f"Seeding submissions into {submissions_dir}/")
    for problem_id, lang, source_filename, source in SUBMISSIONS:
        write_submission(submissions_dir, problem_id, lang, source_filename, source)

    print("Done. Run syncer to load this into Postgres.")


if __name__ == "__main__":
    main()