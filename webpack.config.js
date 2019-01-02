const path = require('path');
const webpack = require('webpack');
const HtmlPlugin = require('html-webpack-plugin');
const CopyWebpackPlugin = require('copy-webpack-plugin');
const { VueLoaderPlugin } = require('vue-loader');

module.exports = {
	entry: './public/main.ts',
	mode: 'development',
	output: {
		path: path.resolve(__dirname, './dist'),
		filename: 'build.js'
	},
	resolve: {
		extensions: ['.js', '.vue', '.json', '.ts'],
		symlinks: false,
		alias: {
			'vue$': 'vue/dist/vue.esm.js',
			'@': path.resolve('./public')
		}
	},
	module: {
		rules: [
			{
				test: /\.vue$/,
				loader: 'vue-loader'
			},
			{
				test: /\.(ts|tsx)?$/,
				exclude: /node_modules/,
				use: [
					{
						loader: 'ts-loader',
						options: {
							// Needed for <script lang="ts"> to work in *.vue files; see https://github.com/vuejs/vue-loader/issues/109
							appendTsSuffixTo: [ /\.vue$/ ]
						}
					},
					{
						loader: 'tslint-loader'
					}
				]
			},
			// {
			// 	test: /\.vue.(ts|tsx)$/,
			// 	exclude: /node_modules/,
			// 	enforce: 'pre',
			// 	use: ['tslint-loader']
			// },
			{
				test: /\.css$/,
				use: [
					'style-loader',
					'css-loader',
					'postcss-loader'
				]
			},
			{
				test: /images\/.*\.(png|jpg|jpeg|gif|svg)$/,
				loader: 'file-loader?name=images/[name].[ext]'
			},
			{
				test: /graphs\/.*\.(gml)$/,
				loader: 'file-loader?name=graphs/[name].[ext]'
			},
			{
				test: /fonts\/.*\.(ttf|otf|eot|svg|woff(2)?)(\?[a-z0-9]+)?$/,
				loader: 'file-loader?name=fonts/[name].[ext]'
			}
		]
	},
	plugins: [
		new VueLoaderPlugin(),
		new HtmlPlugin({
			template: './public/templates/index.template.ejs',
			inject: 'body'
		}),
		new webpack.ProvidePlugin({
			$: 'jquery',
			jQuery: 'jquery'
		}),
		new CopyWebpackPlugin([
			{ from: 'public/static' }
		]),
		new CopyWebpackPlugin([
			{ from: 'public/assets/favicons', to: 'favicons' }
		])
	],
	devtool: '#source-map-eval'
};
