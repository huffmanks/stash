// pnpm i -D eslint typescript-eslint

import { defineConfig, globalIgnores } from "eslint/config";
import tseslint from "typescript-eslint";

const eslintConfig = defineConfig([
  ...tseslint.configs.recommended,
  {
    rules: {
      "@typescript-eslint/no-unused-vars": [
        "warn",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
        },
      ],
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/consistent-type-imports": [
        "error",
        { prefer: "type-imports", fixStyle: "inline-type-imports" },
      ],
    },
  },
  globalIgnores([
    // 1. Universal Build & Outputs
    "dist/**",
    "build/**",
    "out/**",
    ".output/**",
    // 2. Static Assets
    "public/**",
    "static/**",
    // 3. Frameworks & Bundlers
    ".astro/**",
    ".next/**",
    "next-env.d.ts",
    ".vite/**",
    ".tanstack/**",
    ".svelte-kit/**",
    ".expo/**",
    "ios/**",
    "android/**",
    // 4. Tooling
    "node_modules/**",
    "**/*.gen.ts",
    "**/*.gen.tsx",
    ".turbo/**",
    ".netlify/**",
    ".wrangler/**",
    // 5. Environment & Logs
    "logs/**",
    "*.log",
    ".fallow/**",
  ]),
]);

export default eslintConfig;
