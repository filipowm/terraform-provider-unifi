# Code Generation for UniFi Go SDK

The UniFi Go SDK uses a code generation process to create the client code, data models, and REST methods from JSON API 
specifications. This documentation explains how to run the code generator, available parameters, and how to customize 
the output using the `customizations.yml` file.

## Overview

The code generator reads the UniFi Controller API specifications (typically extracted from the UniFi Controller JAR 
or JSON files) and generates Go code that includes data models, REST methods, and validation logic. The generation 
process uses templates (such as `api.go.tmpl` or `client.go.tmpl`) and configuration files.

## Running the Code Generator

There are two common ways to run the code generator from the command line:

### 1. Using `go generate`

The recommended way to regenerate the client code is to use the `go generate` command. This command looks for special 
comments in the source code (usually in a file such as `unifi/codegen.go`) and runs the generator as specified.

Simply navigate to the root of the project and execute:

```bash
go generate unifi/codegen.go
```

This command will:

- Parse the API specifications
- Apply the templates to generate updated models and REST methods
- Overwrite the generated files with the latest code

### 2. Running the Generator Directly with `go run`

Alternatively, you can run the code generator directly using the `go run` command. This approach bypasses `go generate` 
and directly executes the generator file. For example:

```bash
go run ./codegen [OPTIONS] version
```

This command performs the same actions as the `go generate` command, but allows you to customize passed parameters as 
described in the next section.

## Options

The code generator accepts the following flags, which control its behavior. These parameters can be set via command-line 
flags when running the generator, either using `go run ./codegen` or through `go generate` (if configured accordingly):

| Flag                | Description                                                                | Default Value |
|---------------------|----------------------------------------------------------------------------|---------------|
| `-version-base-dir` | The base directory for version JSON files                                  | `.`           |
| `-output-dir`       | The output directory of the generated Go code                              | `.`           |
| `-download-only`    | Only download and build the API structures JSON directory; do not generate |               |
| `-debug`            | Enable debug logging                                                       |               |
| `-trace`            | Enable trace logging (takes precedence over `-debug` flag)                 |               |

*Note:* These flags are defined in the main function of the code generator (see `codegen/main.go`). For further details, 
consult the source code comments in that file.

## Customizing Code Generation with customizations.yml

The `customizations.yml` file is used to fine-tune the code generation process. It can be used to:

- **Override Default Mappings**: Change how certain API fields or types are mapped to Go types.
- **Exclude or Modify Endpoints**: Disable generation for specific API endpoints or alter their behavior.

### How to Use customizations.yml

1. **Edit the File**: Open `codegen/customizations.yml` in your favorite editor and modify the settings according to your needs. 
   The file typically contains sections for field mappings, endpoint customizations, and additional template directives.

2. **Run the Generator**: When you run the generator (using either `go generate` or `go run ./codegen`), 
   it will read the customizations and apply them to the generated code.

3. **Review the Output**: Check the generated files to ensure that your customizations have been applied as expected. 
   Adjust the `customizations.yml` file and re-run the generator if further changes are needed.