const colors = require('tailwindcss/colors')
module.exports = {
  mode: 'jit',
  content: [
    './**/*.{jsx,tsx,html}',
  ],
  darkMode: "class",
  theme: {
    colors: {
      ...colors,
      gray: colors.neutral,
    }
  },
  variants: {},
  plugins: [
    require('@tailwindcss/typography'),
    require("daisyui")
  ],
}