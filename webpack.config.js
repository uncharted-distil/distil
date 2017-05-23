const path = require('path');
const HtmlPlugin = require('html-webpack-plugin');

module.exports = {
    entry: './public/scripts/distil_server.js',
    output: {
        path: path.resolve(__dirname, './dist'),
        filename: 'distil-server.bundle.js'
    },
    module: {
        rules: [
            {
                test: /\.js/,
                exclude: /node_modules/,
                use: 'babel-loader'
                
            },
            {
                test: /\.css$/,
                exclude: /node_modules/,
                use: ['style-loader','css-loader']
            }
        ] 
    },
    plugins: [
        new HtmlPlugin({
            template: './public/templates/index.template.ejs',
            inject: 'body'
        })
    ]
};
