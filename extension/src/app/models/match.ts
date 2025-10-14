export enum SubmissionStatus {
  Accepted = "Accepted",
  CompileError = "Compile Error",
  MemoryLimitExceeded = "Memory Limit Exceeded",
  RuntimeError = "Runtime Error",
  TimeLimitExceeded = "Time Limit Exceeded",
  WrongAnswer = "Wrong Answer",
}

export enum LanguageType {
  C = "c",
  Cpp = "cpp",
  Csharp = "csharp",
  Java = "java",
  Python = "python",
  Python3 = "python3",
  Javascript = "javascript",
  Typescript = "typescript",
  Php = "php",
  Swift = "swift",
  Kotlin = "kotlin",
  Dart = "dart",
  Go = "go",
  Ruby = "ruby",
  Scala = "scala",
  Rust = "rust",
  Racket = "racket",
  Erlang = "erlang",
  Elixir = "elixir",
}

export enum Difficulty {
    Easy = "Easy",
    Medium = "Medium",
    Hard = "Hard",
}

export interface Problem {
    title: string;
    slug: string;
    difficulty: Difficulty;
    tags: number[];
}

export interface MatchDetails {
    isRated: boolean;
    difficulties: Difficulty[];
    tags: number[];
}

export interface PlayerSubmission {
    submissionID: number;
    playerID: number;
    problemID: number;
    status: SubmissionStatus;
}

export interface Session {
    sessionID: string;
    status: string; // "Active", "Won", "Canceled", "Reverted"
    rated: boolean;
    problem: Problem;
    players: number[];
    submissions: PlayerSubmission[];
    winner: number; // <= 0 if no winner
    startTime: string; // ISO 8601 date string
    endTime: string; // ISO 8601 date string
}
