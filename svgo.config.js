module.exports = {
  multipass: true,
  plugins: [
    {
      name: 'preset-default',
      params: {
        overrides: {
          // Keep viewBox for responsive scaling
          removeViewBox: false,
          // Keep IDs if they're needed for styling
          cleanupIds: {
            minify: false
          }
        }
      }
    },
    // Remove unnecessary attributes
    'removeXMLNS',
    // Optimize paths
    {
      name: 'convertPathData',
      params: {
        floatPrecision: 2,
        transformPrecision: 2
      }
    },
    // Merge paths where possible
    'mergePaths',
    // Remove empty containers
    'removeEmptyContainers',
    // Remove unnecessary groups
    'removeUselessStrokeAndFill',
    // Optimize colors
    {
      name: 'convertColors',
      params: {
        currentColor: true
      }
    }
  ]
};