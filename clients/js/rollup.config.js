import resolve from "@rollup/plugin-node-resolve";
import commonjs from "@rollup/plugin-commonjs";
import { terser } from "rollup-plugin-terser";

export default {
  input: "dist/pgrest-client.js",
  output: [
    {
      file: "dist/pgrest-client.esm.js",
      format: "esm",
      sourcemap: true,
    },
    {
      file: "dist/pgrest-client.amd.js",
      format: "amd",
      name: "PGRestClient",
      sourcemap: true,
    },
  ],
  plugins: [resolve(), commonjs(), terser()],
};
