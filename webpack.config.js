const path = require('path');
const webpack = require('webpack');
const HtmlPlugin = require('html-webpack-plugin');
const CopyWebpackPlugin = require('copy-webpack-plugin');

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
				test: /\.(js|vue)$/,
				exclude: /node_modules/,
				enforce: 'pre',
				use: ['eslint-loader']
			},
			{
				test: /\.css$/,
				use: [
					'style-loader',
					'css-loader',
					'postcss-loader'
				]
			},
			{
				test: /\.vue$/,
				loader: 'vue-loader'
			},
			{
				test: /\.js$/,
				exclude: /node_modules/,
				loader: 'babel-loader'
			},
			{
				test: /\.tsx?$/,
				exclude: /node_modules/,
				loader: 'ts-loader',
				options: {
					appendTsSuffixTo: [/\.vue$/]
				}
			},
			{
				test: /images\/.*\.(png|jpg|jpeg|gif|svg)$/,
				loader: 'file-loader?name=images/[name].[ext]'
			},
			{
				test: /fonts\/.*\.(ttf|otf|eot|svg|woff(2)?)(\?[a-z0-9]+)?$/,
				loader: 'file-loader?name=fonts/[name].[ext]'
			}
		]
	},
	plugins: [
		// generates index.html based on generated bundle
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
	devtool: 'source-map-eval'
};
