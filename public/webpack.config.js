
var webpack = require('webpack');
var path = require('path');

var BUILD_DIR = path.resolve(__dirname, 'assets');
var APP_DIR = path.resolve(__dirname, '');

var env = process.env.NODE_ENV;
var config = {
  entry: './index.jsx',

  module: {
    loaders: [
      {
        test: /\.jsx$/,
        loaders: ['babel'],
        exclude: /node_modules/
      },
      {
        test: /\.scss$/,
        loaders: ['style', 'css', 'sass']
      },
      {
        test: /\.(jpg|png|gif)$/,
        loader: 'url-loader'
      }
    ]
  },

  output: {
    path: './',
    filename: 'bundle.js'
  },


  plugins: [
    new webpack.DefinePlugin({
      'process.env.NODE_ENV': JSON.stringify(env)
    }),
    new webpack.optimize.OccurrenceOrderPlugin(),
  ],
  resolve: {
    extensions: ['', '.json', '.js', '.jsx'],
  },
  externals: {
    'marked': 'marked'
  }
};

  config.plugins.push(new webpack.optimize.UglifyJsPlugin({
    compress: {
      warnings: false
    }
  }));


module.exports = config;