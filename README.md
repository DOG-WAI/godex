# godex

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Status](https://img.shields.io/badge/Status-Active-success)]()

## 🏢 Sponsors

Thank you to the following sponsors for supporting the development of godex:

[![Powered by DartNode](https://dartnode.com/branding/DN-Open-Source-sm.png)](https://dartnode.com "Powered by DartNode - Free VPS for Open Source")

## 📋 Project Overview

`godex` is a Go application framework designed for high performance and ease of use.

### Core Features

- 🛡️ **Real-time Detection**: Check if a domain is active.
- 🚀 **High-Performance Caching**: Load data into memory for millisecond response times.
- ⏰ **Scheduled Updates**: Automatically refresh lists from third-party data sources.
- 🎯 **Type Safety**: Utilize a statically declared API with compile-time type checks.
- 📊 **Task Scheduling**: Execute both scheduled and one-time tasks.
- 🔧 **Modular Architecture**: Operate in dual modes (Web/CLI) with optional components.

## 🚀 Quick Start

### Environment Requirements

- Go 1.24+ (for compilation)

### Installation and Running

To get started, follow these steps:

```bash
# Clone the repository
git clone <repository-url>
cd godex

# Install dependencies
go mod tidy

# Start the service
go run cmd/main.go

# Alternatively, build and run
make build
./app
```

### Configuration Files

The application loads configuration files in the following order of priority:

1. `.config.yaml` (Project root directory)
2. `./conf/config.yaml` (Recommended)
3. `/etc/conf/config.yaml` (System-level)

## ⚙️ Configuration File Explanation

The project uses YAML format for configuration files. Below is a sample configuration. For a complete configuration, check `./conf/config-example.yaml`:

```yaml
### System Configuration ###
system:
  show-conf: false  # Set to true for development
```

## 📦 Releases

For the latest releases, visit the [Releases section](https://github.com/DOG-WAI/godex/releases). Here, you can find downloadable files that you can execute.

## 🛠️ Contributing

We welcome contributions to `godex`. If you want to help, please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and commit them.
4. Push your branch to your forked repository.
5. Open a pull request to the main repository.

### Code of Conduct

We expect all contributors to adhere to our [Code of Conduct](CODE_OF_CONDUCT.md).

## 📝 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 📚 Documentation

For detailed documentation, refer to the [Wiki](https://github.com/DOG-WAI/godex/wiki). Here, you can find guides, tutorials, and API references.

## 🔗 Links

- [GitHub Repository](https://github.com/DOG-WAI/godex)
- [Issues](https://github.com/DOG-WAI/godex/issues)
- [Discussions](https://github.com/DOG-WAI/godex/discussions)

## 🤝 Community

Join our community to discuss features, report issues, or seek help. You can find us on:

- [Gitter](https://gitter.im/DOG-WAI/godex)
- [Discord](https://discord.gg/your-invite-link)

## 📅 Roadmap

We have exciting plans for the future of `godex`. Here are some features we aim to implement:

- Enhanced logging capabilities.
- Support for additional data sources.
- Improved user interface for the web application.
- Integration with popular cloud services.

## 🏗️ Architecture

`godex` is built with a modular architecture that allows you to enable or disable components as needed. This flexibility ensures that you can tailor the application to your specific requirements.

### Components

1. **Core**: The main framework that handles requests and responses.
2. **Cache**: Manages in-memory data storage for quick access.
3. **Scheduler**: Handles task execution based on time intervals or specific triggers.
4. **API**: Provides endpoints for interaction with the application.

## 🔍 Testing

We encourage you to run tests to ensure the application works as expected. To run the tests, execute the following command:

```bash
go test ./...
```

You can add your own tests in the `tests` directory.

## 📊 Performance

`godex` is designed for high performance. The caching mechanism ensures that frequently accessed data is readily available, reducing response times. Our benchmarks show that `godex` can handle thousands of requests per second, making it suitable for high-traffic applications.

## 🌐 Internationalization

To support a global audience, `godex` includes features for internationalization. You can easily add new languages by providing translation files in the `locales` directory.

## 📈 Analytics

Integrate analytics to monitor application performance and user engagement. `godex` supports various analytics services. You can configure these in your YAML configuration file.

## 🔒 Security

Security is a priority for `godex`. We implement best practices to protect user data and prevent unauthorized access. Regular updates ensure that any vulnerabilities are promptly addressed.

## 📞 Support

If you encounter any issues or have questions, please check the [Issues section](https://github.com/DOG-WAI/godex/issues) or reach out through our community channels.

## 🎉 Acknowledgments

We would like to thank the contributors and community members who have helped shape `godex`. Your support and feedback are invaluable.

## 🌟 Future Enhancements

We are continuously looking to improve `godex`. Here are some potential enhancements:

- Integration with machine learning libraries for advanced data processing.
- A user-friendly dashboard for managing application settings.
- Enhanced security features, including two-factor authentication.

## 📅 Upcoming Events

Stay tuned for upcoming events, including webinars and community meetups. Check our [Discussions section](https://github.com/DOG-WAI/godex/discussions) for announcements.

## 🔗 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [YAML Documentation](https://yaml.org/spec/1.2/spec.html)
- [Open Source Guides](https://opensource.guide/)

For more information, visit the [Releases section](https://github.com/DOG-WAI/godex/releases) to download the latest version of `godex`.