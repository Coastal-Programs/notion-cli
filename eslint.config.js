// @ts-check
import js from '@eslint/js'
import tseslint from 'typescript-eslint'
import prettier from 'eslint-config-prettier'
import unicornPlugin from 'eslint-plugin-unicorn'
import nodePlugin from 'eslint-plugin-n'
import mochaPlugin from 'eslint-plugin-mocha'

export default tseslint.config(
  // Base ESLint recommended rules
  js.configs.recommended,

  // TypeScript ESLint recommended rules (non-type-checked)
  ...tseslint.configs.recommended,

  // Ignore patterns - must be before other configs
  {
    ignores: [
      'dist/',
      'node_modules/',
      'lib/',
      'scripts/',
      '**/*.js', // Ignore all JS files
      'coverage/',
    ],
  },

  // Language options for TypeScript source files
  {
    files: ['src/**/*.ts'],
    languageOptions: {
      parserOptions: {
        project: './tsconfig.json',
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },

  // Language options for TypeScript test files
  {
    files: ['test/**/*.ts'],
    languageOptions: {
      parserOptions: {
        project: './tsconfig.test.json',
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },

  // Apply to all TypeScript files
  {
    files: ['**/*.ts'],
    plugins: {
      '@typescript-eslint': tseslint.plugin,
      unicorn: unicornPlugin,
      n: nodePlugin,
      mocha: mochaPlugin,
    },
    rules: {
      // TypeScript rules - match previous config
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
        },
      ],

      // Disable type-checked rules that are too strict for current codebase
      '@typescript-eslint/no-unsafe-assignment': 'off',
      '@typescript-eslint/no-unsafe-member-access': 'off',
      '@typescript-eslint/no-unsafe-argument': 'off',
      '@typescript-eslint/no-unsafe-return': 'off',
      '@typescript-eslint/no-unsafe-call': 'off',
      '@typescript-eslint/no-floating-promises': 'off',
      '@typescript-eslint/require-await': 'off',
      '@typescript-eslint/no-unsafe-enum-comparison': 'off',
      '@typescript-eslint/no-redundant-type-constituents': 'off',
      '@typescript-eslint/prefer-promise-reject-errors': 'off',

      // Style rules from oclif
      'capitalized-comments': 'off',
      'comma-dangle': ['error', 'always-multiline'],
      'default-case': 'off',
      'no-multi-spaces': 'off',
      'curly': 'off',
      'indent': ['error', 2, { SwitchCase: 1, MemberExpression: 0 }],
      'quotes': ['error', 'single', { avoidEscape: true }],
      'semi': ['error', 'never'],

      // Unicorn plugin rules
      'unicorn/prevent-abbreviations': 'off',
      'unicorn/no-await-expression-member': 'off',
      'unicorn/no-null': 'off',
      'unicorn/prefer-module': 'warn',
      'unicorn/no-process-exit': 'off',
      'unicorn/numeric-separators-style': 'off',
      'unicorn/prefer-number-properties': 'off',
      'unicorn/switch-case-braces': 'off',
      'unicorn/no-array-for-each': 'off',
      'unicorn/no-negated-condition': 'off',
      'unicorn/prefer-node-protocol': 'off',
      'logical-assignment-operators': 'off',

      // Node plugin rules
      'n/shebang': 'off',
      'n/no-missing-import': 'off',
      'n/no-extraneous-import': 'off',

      // General rules - match previous config
      'no-console': 'off',
      'no-process-exit': 'off',
      'no-implicit-coercion': 'off',
      'object-curly-spacing': ['error', 'always'],
      'valid-jsdoc': 'off',
      'camelcase': 'off',
      'padding-line-between-statements': 'off',
      '@typescript-eslint/no-unused-expressions': 'off',
    },
  },

  // Test file specific overrides
  {
    files: ['test/**/*.ts'],
    languageOptions: {
      globals: {
        describe: true,
        it: true,
        before: true,
        after: true,
        beforeEach: true,
        afterEach: true,
      },
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
      'unicorn/no-null': 'off',
      '@typescript-eslint/no-unused-expressions': 'off',
    },
  },

  // Prettier config (must be last to override formatting rules)
  prettier,
)
