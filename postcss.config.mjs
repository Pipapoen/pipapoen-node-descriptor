import purgecss from '@fullhuman/postcss-purgecss';
export default {
  plugins: [purgecss({
    content: ['./web/index.html', './web/src/**/*.ts', './web/src/**/*.css'],
    safelist: {
      standard: ['active', 'show', 'disabled', 'light', 'dark', 'theme-dark', 'theme-light'],
      greedy: [/data-bs-theme/, /data-theme/]
    },
    keyframes: true,
    fontFace: true
  })]
};
