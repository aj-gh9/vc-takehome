module.exports = {
    extends: ['@commitlint/config-conventional'],
    plugins: ['commitlint-plugin-function-rules'],
    rules: {
        'body-leading-blank': [1, 'always'],
        'footer-leading-blank': [1, 'always'],
        'header-max-length': [0, 'always', 80],
        'scope-case': [2, 'always', ['upper-case', 'lower-case']],
        'scope-empty': [2, 'never'],
        'scope-enum': [0],
        'function-rules/scope-enum': [
            2,
            'always',
            (parsed) => {
                const re = new RegExp('[A-Za-z0-9]+-[0-9]+')
                if (
                    !parsed.scope ||
                    parsed.scope.match(re)
                ) {
                    return [true];
                }

                return [false, `scope must match jira ticket regex ${re.source}`];
            },
        ],
        'subject-empty': [2, 'never'],
        'subject-full-stop': [2, 'never', '.'],
        'subject-case': [2, 'always', [
            'lower-case', // default
            'upper-case', // UPPERCASE
            'camel-case', // camelCase
            'kebab-case', // kebab-case
            'pascal-case', // PascalCase
            'sentence-case', // Sentence case
            'snake-case', // snake_case
            'start-case' // Start Case
        ]],
        'type-case': [2, 'always', 'lower-case'],
        'type-empty': [2, 'never'],
        'type-enum': [
            2,
            'always',
            [
                'build',
                'chore',
                'ci',
                'docs',
                'feat',
                'fix',
                'perf',
                'refactor',
                'revert',
                'style',
                'test'
            ]
        ]
    }
}
