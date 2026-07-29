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
      // 店舗名・説明文に全角スペース (U+3000) を使うため、文字列とテンプレートリテラル内は許可
      'no-irregular-whitespace': ['error', { skipStrings: true, skipTemplates: true }],
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
      // 以下は eslint-plugin-react-hooks v7 で追加された規則で、既存コードが多数抵触する。
      // App.jsx / Map.jsx / SVGButton.jsx のリファクタ後に個別に有効化すること。
      'react-hooks/set-state-in-effect': 'off',
      'react-hooks/static-components': 'off',
      'react-hooks/immutability': 'off',
      'react-hooks/exhaustive-deps': 'off',
    },
  },
  {
    // データ定義クラス (Point / Rect / Shop) を export するだけでコンポーネントは持たない
    files: ['src/MapPoint.jsx'],
    rules: {
      'react-refresh/only-export-components': 'off',
    },
  },
]
