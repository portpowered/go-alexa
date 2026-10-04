import { mkdir, readdir, unlink } from "node:fs/promises";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

const outputDirectory = fileURLToPath(
  new URL("../../pkg/dependencymodels/.asyncapi-generated/", import.meta.url),
);

await mkdir(outputDirectory, { recursive: true });
for (const name of await readdir(outputDirectory)) {
  if (name.endsWith(".go")) {
    await unlink(join(outputDirectory, name));
  }
}
