// Completion spec for readmarker.
// Kiro CLI loads plain .js specs from its configured Specs folder.

const sourceKeyGenerator = {
  script: ["readmarker", "list"],
  postProcess: (out) =>
    out
      .split("\n")
      .map((line) => line.trim())
      .filter((line) => line.length > 0)
      .map((line) => {
        const [sourceKey, cursor] = line.split(/\s+/);
        return {
          name: sourceKey,
          description: cursor ? `current cursor: ${cursor}` : "source key",
        };
      }),
};

const positionArg = {
  name: "pos",
  description: "Non-negative base-10 cursor position",
};

const dbOption = {
  name: "--db",
  description: "Path to the readmarker database",
  args: {
    name: "path",
    template: "filepaths",
  },
};

const helpOptions = [
  { name: ["-h", "--help"], description: "Show help information" },
];

const completionSpec = {
  name: "readmarker",
  description: "Track read cursors for source-agnostic agent workflows",
  subcommands: [
    {
      name: "get",
      description: "Print the cursor for a source key",
      args: {
        name: "source_key",
        description: "Source key to read",
        generators: sourceKeyGenerator,
      },
    },
    {
      name: "advance",
      description: "Advance a source key cursor without moving it backwards",
      args: [
        {
          name: "source_key",
          description: "Source key to advance",
          generators: sourceKeyGenerator,
        },
        positionArg,
      ],
    },
    {
      name: "set",
      description: "Set a source key cursor to an exact position",
      args: [
        {
          name: "source_key",
          description: "Source key to set",
          generators: sourceKeyGenerator,
        },
        positionArg,
      ],
    },
    {
      name: "list",
      description: "List all stored source keys and cursors",
      options: [
        {
          name: "--json",
          description: "Emit JSON output",
        },
        ...helpOptions,
      ],
    },
    {
      name: "help",
      description: "Show readmarker help",
    },
  ],
  options: [
    dbOption,
    { name: "--version", description: "Print version information" },
    ...helpOptions,
  ],
};

export default completionSpec;
