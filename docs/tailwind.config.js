const colors = require('tailwindcss/colors')
const defaultTheme = require('tailwindcss/defaultTheme')

module.exports = {
  mode: 'jit',
  content: [
    './**/*.{jsx,tsx,html,mdx}',
  ],
  // darkMode: "class",
  theme: {
    fontFamily: {
      ...defaultTheme.fontFamily,
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
  daisyui: {
    themes: [{
      light1: {
        "primary": "#3D59C7",
        "primary-content":"#fff",
        "secondary": "#8a89d6",
        "accent": "#045c70",
        "neutral": "#FFFFFF",
        "base-100": "#fafafa",
        "base-200": "#e5e5e5",
        "info": "#bfdbf8",
        "success": "#25B675",
        "warning": "#BA6E03",
        "error": "#F45D88",
      },
    }, {
      dark1: {
        "primary": "#3D59C7",
        "secondary": "#6cd8d5",
        "accent": "#d343a8",
        "neutral": "#242526",
        "base-100": "#1B1B1D",
        "base-200": "#262626",
        "info": "#3D59C7",
        "success": "#5DEAA8",
        "warning": "#F9A01A",
        "error": "#EC556E",
      }
    }],
    darkTheme: "dark1",
  },
}