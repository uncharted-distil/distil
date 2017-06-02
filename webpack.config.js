const path = require('path');
const HtmlPlugin = require('html-webpack-plugin');

module.exports = {
    entry: './public/main.js',
    output: {
        path: path.resolve(__dirname, './dist'),
        filename: 'build.js'
    },
    resolve: {
        extensions: ['.js', '.vue', '.json'],
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
                test: /\.vue$/,
                loader: 'vue-loader'
            },
            {
                test: /\.js$/,
                exclude: /node_modules/,
                use: ['babel-loader']
            },
            {
                test: /\.(png|jpg|gif|svg)$/,
                loader: 'file-loader',
                options:  {
                    name: '[name].[ext]?[hash]'
                }
            }
        ] 
    },
    plugins: [
        // generates index.html based on generated bundle
        new HtmlPlugin({
            template: './public/templates/index.template.ejs',
            inject: 'body'
        })
    ],
    devtool: 'source-map'
};
