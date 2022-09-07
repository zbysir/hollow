const colors = require('tailwindcss/colors')
const defaultTheme = require('tailwindcss/defaultTheme')

module.exports = {
  mode: 'jit',
  content: [
    './**/*.{jsx,tsx,html}',
  ],
  // darkMode: "class",
  theme: {
    colors: {
      ...colors,
      gray: colors.neutral,
    },
    fontFamily: {
      ...defaultTheme.fontFamily,
      'noto': ['Noto Serif SC', 'serif'],
    }
  },
  variants: {},
  plugins: [
    require('@tailwindcss/typography')({
      // :where 在手机上兼容性不佳，不启用
      // https://github.com/tailwindlabs/tailwindcss-typography/pull/203
      target: 'legacy'
    }),
    require("daisyui")
  ],
}