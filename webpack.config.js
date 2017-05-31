const path = require('path');
const HtmlPlugin = require('html-webpack-plugin');

module.exports = {
    entry: './public/main.js',
    output: {
        path: path.resolve(__dirname, './dist'),
        filename: 'build.js'
    },
    module: {
        rules: [
            {
                test: /\.js/,
                exclude: /node_modules/,
                use: ['babel-loader', 'eslint-loader']
            },
            {
                test: /\.vue/,
                loader: 'vue-loader'
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
    resolve: {
        alias: {
        'vue$': 'vue/dist/vue.esm.js'
        }
    },
    plugins: [
        // generates index.html based on generated bundle
        new HtmlPlugin({
            template: './public/templates/index.template.ejs',
            inject: 'body'
        })
    ]
};
