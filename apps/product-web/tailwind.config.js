/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        monti: {
          bg: '#050814',
          panel: '#0c1425',
          line: 'rgb(91 120 177 / 24%)',
          cyan: '#16c7ff',
          blue: '#2375ff',
          violet: '#8d39ff',
          muted: '#8390aa',
          danger: '#ff5c7a',
          success: '#3dd68c'
        }
      }
    }
  },
  plugins: []
};
