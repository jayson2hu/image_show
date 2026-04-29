/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        ink: '#17202a',
        mist: '#eef4f7',
        coral: '#e56f5a',
        teal: '#177e89',
      },
    },
  },
  plugins: [],
}
