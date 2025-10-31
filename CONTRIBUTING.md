# Contributing to MCPCan

Thank you for your interest in contributing to MCPCan! We welcome contributions from everyone, whether you're fixing bugs, adding features, improving documentation, or helping with translations.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [opensource@kymo.cn](mailto:opensource@kymo.cn).

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+** for backend development
- **Node.js 18+** and **pnpm** for frontend development
- **Docker** and **Docker Compose** for containerization
- **Kubernetes** (minikube, k3s, or full cluster) for deployment testing
- **Git** for version control

### Fork and Clone

1. Fork the MCPCan repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/mcpcan.git
   cd mcpcan
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/kymo-mcp/mcpcan.git
   ```

## Development Setup

### Backend Setup

```bash
cd backend
go mod download
cp config/gateway.yaml.example config/gateway.yaml
# Edit configuration files as needed
go run cmd/gateway/main.go
```

### Frontend Setup

```bash
cd web
pnpm install
cp .env.example .env.local
# Edit environment variables as needed
pnpm dev
```

### Full Stack Development

Use Docker Compose for full stack development:

```bash
docker-compose -f docker-compose.dev.yml up
```

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check the [issue tracker](https://github.com/kymo-mcp/mcpcan/issues) to see if the problem has already been reported.

When creating a bug report, please include:

- **Clear title and description**
- **Steps to reproduce** the behavior
- **Expected behavior**
- **Actual behavior**
- **Screenshots** if applicable
- **Environment details** (OS, browser, versions)
- **Additional context** or logs

### Suggesting Enhancements

Enhancement suggestions are welcome! Please provide:

- **Clear title and description**
- **Use case** and **motivation**
- **Detailed explanation** of the proposed feature
- **Possible implementation** approach
- **Alternative solutions** considered

### Contributing Code

1. **Check existing issues** or create a new one to discuss your changes
2. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes** following our coding standards
4. **Add tests** for new functionality
5. **Update documentation** if needed
6. **Commit your changes** with clear commit messages
7. **Push to your fork** and create a pull request

## Pull Request Process

### Before Submitting

- [ ] Code follows the project's coding standards
- [ ] Tests pass locally
- [ ] Documentation is updated
- [ ] Commit messages are clear and descriptive
- [ ] Branch is up to date with `main`

### PR Guidelines

1. **Title**: Use a clear, descriptive title
2. **Description**: Explain what changes you made and why
3. **Link issues**: Reference related issues using `Fixes #123`
4. **Screenshots**: Include screenshots for UI changes
5. **Testing**: Describe how you tested your changes

### Review Process

- All PRs require at least one review from a maintainer
- Address feedback promptly and professionally
- Keep PRs focused and reasonably sized
- Be patient - reviews take time

## Coding Standards

### Go Backend

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` and `golint`
- Write meaningful variable and function names
- Add comments for exported functions and complex logic
- Handle errors appropriately

```go
// Good
func GetUserByID(id string) (*User, error) {
    if id == "" {
        return nil, errors.New("user ID cannot be empty")
    }
    // Implementation...
}
```

### TypeScript Frontend

- Follow [TypeScript ESLint rules](https://typescript-eslint.io/)
- Use meaningful component and variable names
- Prefer functional components with hooks
- Add TypeScript types for all props and state

```typescript
// Good
interface UserProfileProps {
  user: User;
  onUpdate: (user: User) => void;
}

const UserProfile: React.FC<UserProfileProps> = ({ user, onUpdate }) => {
  // Implementation...
};
```

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
type(scope): description

[optional body]

[optional footer]
```

Examples:
- `feat(auth): add OAuth2 authentication`
- `fix(api): handle null response in user endpoint`
- `docs(readme): update installation instructions`
- `test(user): add unit tests for user service`

## Testing

### Backend Testing

```bash
cd backend
go test ./...
go test -race ./...
go test -cover ./...
```

### Frontend Testing

```bash
cd web
pnpm test
pnpm test:coverage
pnpm test:e2e
```

### Integration Testing

```bash
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## Documentation

### Code Documentation

- Document all exported functions and types
- Use clear, concise comments
- Include examples for complex functions
- Keep documentation up to date with code changes

### User Documentation

- Update README.md for new features
- Add or update API documentation
- Include configuration examples
- Write clear installation and usage instructions

## Community

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Discord**: Real-time chat and community support
- **Email**: [opensource@kymo.cn](mailto:opensource@kymo.cn)

### Getting Help

- Check existing documentation and issues first
- Ask questions in GitHub Discussions
- Join our Discord community
- Attend community meetings (schedule TBD)

## Recognition

Contributors will be recognized in:

- Project README.md
- Release notes for significant contributions
- Annual contributor appreciation posts
- Special contributor badges (coming soon)

## License

By contributing to MCPCan, you agree that your contributions will be licensed under the [GPL-3.0 License](LICENSE).

---

Thank you for contributing to MCPCan! Your efforts help make this project better for everyone. 🚀