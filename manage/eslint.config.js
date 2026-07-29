import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'

export default [
  { ignores: ['dist'] },
  js.configs.recommended,
  reactHooks.configs.flat['recommended-latest'],
  {
    files: ['**/*.{js,jsx}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: globals.browser,
      parserOptions: {
        ecmaFeatures: { jsx: true },
      },
    },
    plugins: {
      'react-refresh': reactRefresh,
    },
    rules: {
      // 管理画面の文言に全角スペース (U+3000) を使うため、文字列とテンプレートリテラル内は許可
      'no-irregular-whitespace': ['error', { skipStrings: true, skipTemplates: true }],
      // 使用しないコールバック引数は _ 始まりで明示する
      'no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
      // 以下は eslint-plugin-react-hooks v7 で追加された規則で、既存の
      // 「マウント時にフェッチして setState」パターンが抵触する。
      // 各ページをリファクタしたら個別に有効化すること。
      'react-hooks/set-state-in-effect': 'off',
      'react-hooks/immutability': 'off',
    },
  },
]
