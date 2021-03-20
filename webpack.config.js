const { CleanWebpackPlugin } = require('clean-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin')
const glob = require('glob')
const path = require("path");
const webpack = require("webpack");

let pages = glob.sync(path.resolve(__dirname, 'ui/pages/**/*.jsx')).reduce((acc, path) => {
	const entry = path.replace(new RegExp('^.+/ui/pages/'), '').replace(".jsx", '');
	if(entry.match(/index$/)) {
		acc.push(entry + '.html')
	} else {
		acc.push(entry + '/index.html')
	}

	return acc
}, []);

let mode = 'development';
let outPath = '/content/';
if (process.env.production) {
	mode = 'production';
}

module.exports = {
	mode: mode,
	entry: {
		app: path.resolve(__dirname, 'ui/app.jsx'),
	},
	output: {
		filename: '[name].js',
		chunkFilename: '[chunkhash].bundle.js',
		path: path.resolve(__dirname) + outPath,
		publicPath: '/',
	},
	plugins: [
		new CleanWebpackPlugin({
			cleanOnceBeforeBuildPatterns: ['**/*', '!CNAME'],
		}),
		new webpack.EnvironmentPlugin({
			production: false,
		}),
		...pages.map(page => new HtmlWebpackPlugin({
			title: 'Jabberwocky',
			filename: page,
		})),
	],
	context: path.resolve(__dirname),
	resolve: {
		extensions: ["*", ".js", ".jsx"],
		modules: [path.resolve(__dirname, 'node_modules')],
		alias: {
			Page: path.resolve(__dirname, 'ui/pages/'),
			Component: path.resolve(__dirname, 'ui/components/'),
			Include: path.resolve(__dirname, 'ui/includes/'),
		}
	},
	optimization: {
		minimize: true,
		usedExports: true,
		runtimeChunk: 'single',
		moduleIds: 'hashed',
		splitChunks: {
			cacheGroups: {
				react: {
					test: /[\\/]node_modules[\\/](react|react-dom)[\\/]/,
					name: 'react',
					chunks: 'all',
				},
			},
		},
	},
	module: {
		rules: [
			{
				test: /\.(js|jsx)$/,
				exclude: /(node_modules|bower_components)/,
				loader: "babel-loader",
				options: {
					presets: ["@babel/env"],
					"plugins": ["minify-dead-code-elimination"],
				}
			},
			{
				test: /\.css$/,
				use: ["style-loader", "css-loader"]
			}
		]
	},
	devServer: {
		contentBase: path.join(__dirname, 'dist'),
		compress: true,
		disableHostCheck: true,   // That solved it
		port: 9004
	},
};
