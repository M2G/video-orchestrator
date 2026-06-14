import { createRequire } from "node:module";
import { join } from "node:path";
import type { IncomingMessage } from "node:http";

const require = createRequire(import.meta.url);
const native = require(join(process.cwd(), "build/Release/multipart.node"));