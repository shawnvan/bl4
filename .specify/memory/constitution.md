<!-- Sync Impact Report:
Version change: 0.0.0 → 1.0.0
Modified principles: N/A (initial constitution)
Added sections: Core Principles (5), Development Workflow, Quality Assurance, Governance
Removed sections: N/A
Templates requiring updates: ✅ plan-template.md (Constitution Check), ✅ tasks-template.md (TDD alignment), ✅ spec-template.md (requirements alignment), ⚠️ agent-file-template.md (general guidance), ⚠️ checklist-template.md (quality gates)
Follow-up TODOs: None
-->

# BL4 Constitution

## Core Principles

### I. 高质量代码标准 (High-Quality Code Standards)

必须保持高标准的代码质量、可读性和可维护性。要求清晰的代码结构、良好的文档、一致的格式，函数和组件必须单一职责且具有有意义的命名。代码审查必须确保遵循既定模式和最佳实践。

**Rationale**: High-quality code is the foundation of maintainable software. This principle ensures long-term project viability and reduces technical debt.

### II. 用户体验一致性 (User Experience Consistency)

所有功能和交互的用户界面和体验必须保持一致。设计模式、配色方案、排版和交互行为必须遵循既定的设计系统，只有在功能规范明确要求时才允许变化。

**Rationale**: Consistent user experience reduces cognitive load and creates intuitive interactions that build user trust and satisfaction.

### III. 测试驱动开发（不可协商）(Test-Driven Development - Non-negotiable)

所有功能都必须采用测试驱动开发。必须严格遵循红绿重构循环：先写失败的测试 → 让测试通过 → 重构代码。必须包含全面的测试覆盖，包括单元测试和集成测试。

**Rationale**: TDD ensures code quality, provides regression safety, and forces thoughtful design before implementation.

### IV. 最小需求原则 (Minimum Requirements Principle)

必须严格按照规范实现需求，不添加不必要的功能或未来扩展性。不包含明确需求之外的额外功能。如有模糊性或额外功能看似有益，必须事先获得确认。

**Rationale**: Prevents scope creep and ensures focused delivery of actual user value without unnecessary complexity.

### V. 技术文档验证（不可协商）(Technical Documentation Verification - Non-negotiable)

在制定技术方案、选择第三方库时，必须使用MCP查询最新官方文档。所有函数调用、API使用模式和库集成必须根据当前文档进行验证。必须在编写实现代码之前对任何技术决策进行文档查询。

**Rationale**: Ensures technical decisions are based on current, accurate information and prevents integration issues from outdated documentation assumptions.

## Development Workflow

### Requirements to Implementation Process

1. **Specification Phase**: Clear user stories with measurable acceptance criteria
2. **Planning Phase**: Technical research and design based on verified documentation
3. **Task Generation**: Dependency-ordered tasks aligned with user story priorities
4. **TDD Implementation**: Red-Green-Refactor cycle strictly enforced
5. **Quality Gates**: Code review and testing validation before merge

### Code Review Requirements

- Must verify compliance with all constitutional principles
- Must ensure TDD cycle was followed (failing tests exist before implementation)
- Must validate technical decisions against current documentation
- Must check for scope adherence (no extra features beyond specification)

## Quality Assurance

### Testing Requirements

- **Unit Tests**: Must cover all business logic and edge cases
- **Integration Tests**: Must verify component interactions and data flow
- **Contract Tests**: Must validate external service integrations
- **User Journey Tests**: Must verify complete user workflows independently

### Code Quality Standards

- Consistent formatting and naming conventions
- Single responsibility principle for all functions and components
- Clear documentation for complex logic
- No hardcoded values without explicit justification

## Governance

### Amendment Process

This constitution represents the non-negotiable standards for BL4 development. Amendments require:

1. **Proposal**: Written proposal with specific changes and rationale
2. **Review**: Impact assessment on existing codebase and workflows
3. **Approval**: Consensus approval from project maintainers
4. **Migration**: Plan for updating existing code and documentation
5. **Communication**: Clear communication of changes to all contributors

### Versioning Policy

- **MAJOR**: Backward incompatible governance or principle changes
- **MINOR**: New principles or expanded guidance sections
- **PATCH**: Clarifications, wording improvements, non-semantic refinements

### Compliance Review

All pull requests and code reviews must explicitly verify compliance with current constitutional standards. Complexity beyond standard patterns must be justified in terms of specific user value delivered.

**Version**: 1.0.0 | **Ratified**: 2025-10-30 | **Last Amended**: 2025-10-30