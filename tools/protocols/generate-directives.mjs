import { GoFileGenerator, GO_COMMON_PRESET } from "@asyncapi/modelina";
import { fileURLToPath } from "node:url";

const source = new URL("../../api/asyncapi.yaml", import.meta.url);
const output = fileURLToPath(
  new URL("../../pkg/dependencymodels/.asyncapi-generated/", import.meta.url),
);
const generator = new GoFileGenerator({
  presets: [{ preset: GO_COMMON_PRESET, options: { addJsonTag: true } }],
});

// Preserve the source URL so external schema references resolve beside the
// AsyncAPI file, independently of the process working directory.
await generator.generateToFiles(source.href, output, {
  packageName: "alexamodels",
});
