# Token budget (always on)
- Grep -n / Glob before Read; Read slices (offset+limit <=150).
- Don't re-read files already in context; don't re-run passing tests.
- Prefer Edit over Write for existing files. Batch independent tool calls in one turn.
- Search/test/log-heavy work -> haiku subagents (`explorer`, `test-runner`); they return <=15 lines.
- Final reply <=8 lines; no code recap; no plan recap; no filler.
- Do not open: Library, Temp, node_modules, bin, obj, *.meta, *.unity, *.prefab, *.asset, logs.
