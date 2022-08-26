module.exports = {
  mode: 'jit',
  content: [
    './**/*.{jsx,tsx,html}',
  ],
  darkMode: "class",
  theme: {},
  variants: {},
  plugins: [
    require('@tailwindcss/typography'),
    require("daisyui")
  ],
}