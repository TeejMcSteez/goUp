import js from "@eslint/js";
import globals from "globals";
import eslintReact from "@eslint-react/eslint-plugin";
import reactHooks from "eslint-plugin-react-hooks";
import tseslint from "typescript-eslint";

export default tseslint.config(
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ["**/*.{ts,tsx}"],
    ...eslintReact.configs.recommended,
    plugins: {
      ...eslintReact.configs.recommended.plugins,
      "react-hooks": reactHooks,
    },
    languageOptions: {
      globals: {
        ...globals.browser,
      },
    },
    rules: {
      ...eslintReact.configs.recommended.rules,
      ...reactHooks.configs.recommended.rules,
      "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
    },
  },
);
