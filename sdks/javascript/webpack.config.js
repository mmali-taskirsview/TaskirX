const path = require('path');

module.exports = {
  entry: './src/index.ts',
  output: {
    filename: 'adx-sdk.js',
    path: path.resolve(__dirname, 'dist'),
    library: {
      name: 'AdxSDK',
      type: 'umd',
      export: 'default'
    },
    globalObject: 'this'
  },
  resolve: {
    extensions: ['.ts', '.js']
  },
  module: {
    rules: [
      {
        test: /\.ts$/,
        use: 'ts-loader',
        exclude: /node_modules/
      }
    ]
  },
  optimization: {
    minimize: true
  }
};
