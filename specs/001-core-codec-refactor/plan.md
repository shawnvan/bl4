# Implementation Plan: Core Codec Algorithm Refactor

**Branch**: `001-core-codec-refactor` | **Date**: 2025-11-03 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-core-codec-refactor/spec.md`
**Additional Requirements**: 保持ui风格和交互方式不变；解码和编码部分算法实现方式和参考项目必须一致，可以以包引入的方式直接引入参考项目

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Refactor the core BL4 item codec algorithms by adopting reference implementation algorithms (https://github.com/Nicnl/borderlands4-serials) while maintaining complete UI/API compatibility. Research shows current implementation already follows best practices; we'll enhance accuracy through VARINT decoder improvements (4-bit nibble processing), optimize Base85 performance, and refine tokenization. Progressive replacement strategy: VARINT decoder → Base85 codec → serialization engine, with 10% performance improvement and 15% memory reduction targets.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Reference project (github.com/Nicnl/borderlands4-serials), existing HTTP server framework, Gin/Fiber for web API
**Storage**: N/A (stateless processing)
**Testing**: Go testing framework + existing test suite
**Target Platform**: Linux/macOS/Windows (CLI + Web API)
**Project Type**: Single project with CLI and web components
**Performance Goals**: 10% faster individual item processing, 15% reduced memory usage for batch processing
**Constraints**: Must maintain 100% API/UI compatibility, preserve all existing endpoints, ensure backward compatibility
**Scale/Scope**: Current codebase size, no new features added

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Required Compliance Gates

- **[GATE]** Technical Documentation Verification: All technology choices and APIs verified using MCP current documentation ✅
  - Reference project algorithms documented and verified
  - Go 1.21+ compatibility verified via official documentation
  - Third-party library integrations verified via current documentation
- **[GATE]** Minimum Requirements Principle: Implementation plan contains only explicitly required features from specification ✅
- **[GATE]** TDD Readiness: Plan includes comprehensive testing strategy (unit, integration, contract tests) ✅
- **[GATE]** Quality Code Standards: Planned structure supports maintainable, well-documented code ✅
- **[GATE]** UX Consistency: Design approach follows established patterns (if UI components involved) ✅

### Post-Design Compliance Check

All gates passed. Research completed, design finalized, ready for implementation phase. Progressive replacement strategy ensures minimal risk while delivering required improvements.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/                    # Application entry points
├── api/               # HTTP API server
├── cli/               # Command-line interface
└── tui/               # Terminal user interface

internal/              # Private application code
├── api/               # REST API layer
│   ├── handlers/      # HTTP request handlers
│   ├── middleware/    # HTTP middleware
│   ├── models/        # API data models
│   └── services/      # Business logic services
├── codec/             # Core codec implementation (REFACTOR TARGET)
│   ├── base85/        # Base85 encoding/decoding (Phase 2)
│   ├── bitstream/     # Bit-level stream processing
│   ├── datatypes/     # VARINT and other data types (Phase 1)
│   ├── serial/        # Serialization engine (Phase 3)
│   ├── token/         # Tokenization engine
│   └── utils/         # Utility functions
├── cli/               # CLI application logic
├── tui/               # Terminal UI components
└── config/            # Configuration management

pkg/                   # Public packages
├── logger/            # Logging utilities
└── validator/         # Validation utilities

tests/                 # Test suite
├── unit/              # Unit tests
├── integration/       # Integration tests
└── contract/          # Contract tests

web/                   # Web UI assets
├── static/            # Static files (CSS, JS)
└── templates/         # HTML templates
```

**Structure Decision**: Single project with multiple interfaces (API, CLI, TUI) sharing common codec core. Refactor focuses on `internal/codec/` packages while maintaining all existing interfaces and contracts.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| No violations found | N/A | N/A |

---

## Phase 0: Research & Analysis (Completed)

### Reference Project Analysis
- **Decision**: Adopt reference implementation algorithms while maintaining Go performance advantages
- **Rationale**: Reference provides proven accuracy improvements; Go implementation offers superior performance
- **Alternatives considered**: Direct Python port (rejected due to performance), C integration (rejected due to complexity)

### Algorithm Implementation Strategy
- **Decision**: Progressive replacement strategy (VARINT → Base85 → Serialization)
- **Rationale**: Minimizes risk, allows thorough validation at each phase
- **Alternatives considered**: Big-bang replacement (rejected due to risk), partial refactoring (rejected due to complexity)

### Integration Approach
- **Decision**: Maintain all existing interfaces, replace only internal algorithms
- **Rationale**: Ensures zero disruption to existing users and integrations
- **Alternatives considered**: API redesign (rejected due to compatibility requirements)

## Phase 1: Design & Contracts (Completed)

### Data Model Design
Entities and relationships defined in `data-model.md` with:
- Clear validation rules
- Backward compatibility guarantees
- Comprehensive error handling

### API Contracts
REST API specifications defined in `contracts/api.yaml` with:
- Complete OpenAPI 3.0 specification
- Full backward compatibility
- Comprehensive error handling

### Quick Start Guide
Developer onboarding created in `quickstart.md` with:
- Step-by-step implementation guidance
- Testing strategies
- Troubleshooting information

### Agent Context Update
Claude Code context updated with:
- Go 1.21+ language requirements
- Reference project dependencies
- Single project architecture details

---

## Final Status

### Generated Artifacts

✅ **Research Report** (`research.md`) - Complete analysis of reference project and implementation strategy
✅ **Data Model** (`data-model.md`) - Comprehensive entity definitions and relationships
✅ **API Contracts** (`contracts/api.yaml`) - Complete OpenAPI specification
✅ **Quick Start Guide** (`quickstart.md`) - Developer implementation guide
✅ **Agent Context** - Updated for Claude Code with new requirements

### Ready for Implementation

- All constitutional gates passed ✅
- Technical decisions verified and documented ✅
- Design artifacts complete and consistent ✅
- TDD strategy defined and ready ✅

**Next Steps**: Execute `/speckit.tasks` to generate detailed implementation tasks, then proceed with `/speckit.implement`.

---

**Implementation Path**: Ready for task generation and development execution with comprehensive guidance and validated technical decisions.