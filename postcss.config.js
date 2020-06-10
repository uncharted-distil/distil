module.exports = {
  plugins: {
    "postcss-import": {},
    "postcss-cssnext": {
      features: {
        customProperties: {
          warnings: false
        }
      }
    },
    cssnano: {
      reduceIdents: false
    }
  }
};

