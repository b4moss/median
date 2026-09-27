#!/usr/bin/env node
import { pathToFileURL } from "node:url";
import { parseArgs } from "node:util";
import { createMediaSQL, ID_AUTO_INCREMENT } from "./db/schema.js";

const USAGE = `median — media library CLI

Usage:
  median migrate dump --dialect <mysql|postgres|sqlite3> [--id-strategy <strategy>] [--table <name>]

Options:
  --dialect       Required. mysql | postgres | sqlite3
  --id-strategy   Default auto_increment. auto_increment | uuid_v4 | uuid_v7 | ulid
  --table         Default media

SQL is written to stdout only. Errors and usage go to stderr.`;

/** CLI entry for tests. Returns process exit code. */
export function run(
  argv: string[],
  stdout: { write(s: string): void } = process.stdout,
  stderr: { write(s: string): void } = process.stderr,
): number {
  if (argv.length === 0) {
    stderr.write(USAGE + "\n");
    return 2;
  }
  const [cmd, ...rest] = argv;
  if (cmd === "help" || cmd === "-h" || cmd === "--help") {
    stderr.write(USAGE + "\n");
    return 0;
  }
  if (cmd !== "migrate") {
    stderr.write(`unknown command "${cmd}"\n\n${USAGE}\n`);
    return 2;
  }
  if (rest.length === 0) {
    stderr.write(`migrate: missing subcommand (want dump)\n\n${USAGE}\n`);
    return 2;
  }
  const [sub, ...dumpArgs] = rest;
  if (sub !== "dump") {
    stderr.write(`migrate: unknown subcommand "${sub}" (want dump)\n\n${USAGE}\n`);
    return 2;
  }
  return runMigrateDump(dumpArgs, stdout, stderr);
}

function runMigrateDump(
  args: string[],
  stdout: { write(s: string): void },
  stderr: { write(s: string): void },
): number {
  let values: {
    dialect?: string;
    "id-strategy"?: string;
    table?: string;
    help?: boolean;
  };
  try {
    const parsed = parseArgs({
      args,
      options: {
        dialect: { type: "string" },
        "id-strategy": { type: "string" },
        table: { type: "string" },
        help: { type: "boolean", short: "h" },
      },
      strict: true,
      allowPositionals: false,
    });
    values = parsed.values;
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e);
    stderr.write(`migrate dump: ${msg}\n\n${USAGE}\n`);
    return 2;
  }
  if (values.help) {
    stderr.write(USAGE + "\n");
    return 0;
  }
  const dialect = (values.dialect ?? "").trim();
  if (!dialect) {
    stderr.write(`migrate dump: --dialect is required\n\n${USAGE}\n`);
    return 2;
  }
  const idStrategy = (values["id-strategy"] ?? ID_AUTO_INCREMENT).trim() || ID_AUTO_INCREMENT;
  const table = values.table;
  try {
    const sql = createMediaSQL(dialect, idStrategy, table).trim();
    stdout.write(sql + "\n");
    return 0;
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e);
    stderr.write(`migrate dump: ${msg}\n`);
    return 1;
  }
}

const isDirect =
  Boolean(process.argv[1]) && import.meta.url === pathToFileURL(process.argv[1]!).href;

if (isDirect) {
  process.exitCode = run(process.argv.slice(2));
}
