module.exports = {
  mode: 'jit',
  content: [
    './**/*.{jsx,tsx,html}',
  ],
  theme: {},
  variants: {},
  plugins: [
    require('@tailwindcss/typography'),
    // ...
  ],
}