module.exports = (_, argv) => {
    const isDev = argv && argv.mode === 'development';
    return {
        mode: isDev ? 'development' : 'production',
        entry: ['./index.js'],
        output: {
            filename: 'bundle.js'
        },
        module: {
            rules: [
                {
                    test: /\.css$/i,
                    use: ['style-loader', 'css-loader'],
                },
            ],
        },
        devtool: isDev ? 'eval-source-map' : false,
        watch: isDev,
        watchOptions: {
            ignored: /node_modules/
        }
    };
};
