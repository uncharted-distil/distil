const path = require("path");
const webpack = require("webpack");
const HtmlPlugin = require("html-webpack-plugin");
const CopyWebpackPlugin = require("copy-webpack-plugin");
const { VueLoaderPlugin } = require("vue-loader");

module.exports = {
  entry: "./public/main.ts",
  mode: "development",
  output: {
    path: path.resolve(__dirname, "./dist"),
    filename: "build.js",
    // Taken from https://www.mistergoodcat.com/post/the-joy-that-is-source-maps-with-vuejs-and-typescript to prevent duplicate sources in chrome debugger
    devtoolModuleFilenameTemplate: (info) => {
      let $filename = "sources://" + info.resourcePath;
      if (
        info.resourcePath.match(/\.vue$/) &&
        !info.query.match(/type=script/)
      ) {
        $filename =
          "webpack-generated:///" + info.resourcePath + "?" + info.hash;
      }
      return $filename;
    },
    devtoolFallbackModuleFilenameTemplate: "webpack:///[resource-path]?[hash]",
  },
  resolve: {
    extensions: [".js", ".vue", ".json", ".ts"],
    symlinks: false,
  },
  module: {
    rules: [
      {
        test: /\.vue$/,
        loader: "vue-loader",
      },
      {
        test: /\.(ts|tsx)?$/,
        exclude: /node_modules/,
        use: [
          {
            loader: "ts-loader",
            options: {
              // Needed for <script lang="ts"> to work in *.vue files; see https://github.com/vuejs/vue-loader/issues/109
              appendTsSuffixTo: [/\.vue$/],
            },
          },
        ],
      },
      {
        test: /\.css$/,
        use: ["style-loader", "css-loader", "postcss-loader"],
      },
      {
        test: /images\/.*\.(png|jpg|jpeg|gif|svg)$/,
        loader: "file-loader?name=images/[name].[ext]",
      },
      {
        test: /graphs\/.*\.(gml)$/,
        loader: "file-loader?name=graphs/[name].[ext]",
      },
      {
        test: /fonts\/.*\.(ttf|otf|eot|svg|woff(2)?)(\?[a-z0-9]+)?$/,
        loader: "file-loader?name=fonts/[name].[ext]",
      },
    ],
  },
  plugins: [
    new VueLoaderPlugin(),
    new HtmlPlugin({
      template: "./public/templates/index.template.ejs",
      inject: "body",
    }),
    new webpack.ProvidePlugin({
      // Required for facets.js peer dependency
      $: "jquery",
      jQuery: "jquery",
    }),
    new CopyWebpackPlugin({
      patterns: [
        { from: "public/static" },
        { from: "public/assets/favicons", to: "favicons" },
      ],
    }),
  ],
  devtool: "eval-source-map",
};
