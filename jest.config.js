module.exports = {
    verbose: true,
    collectCoverageFrom: [
        'server/**/*.{js}',
        '!**/node_modules/**',
        '!server/public/**',
        '!client/**'
    ],
    testMatch: [
        '<rootDir>/server/**/__tests__/**/*.js',
        '<rootDir>/server/**/?(*.)(spec|test).js'
    ],
    testEnvironment: 'node',
    testURL: 'http://localhost',
    moduleFileExtensions: [
        'js',
        'json'
    ]
};
