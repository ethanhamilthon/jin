import { writeFileSync } from "node:fs";

writeFileSync(new URL("../../internal/web/dist/.gitkeep", import.meta.url), "");
