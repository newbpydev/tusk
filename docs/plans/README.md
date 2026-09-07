# Tusk Feature Plan Registry

This directory contains the canonical Brainstorm specification, modular implementation plans, verification plans, and issue workorders for the Tusk task management reboot.

> **Authoritative Orchestration**: The live, real-time checklist tracking active phases, implementation units, and agent execution pointers is maintained at [MASTERPLAN.md](../../MASTERPLAN.md) in the repository root.

## Planning Hierarchy & Execution Sequence

```mermaid
graph TD
    B[Brainstorm: Modern Task System<br/>2026-09-06-001-feat-tusk-modern-task-system-plan.md] --> P0[Plan 000: Setup & Multi-Agent Foundation<br/>Active Foundation - Complete]
    P0 --> P1[Plan 001: Core Domain & Invariants<br/>Ultrathink Planning Pack Ready]
    P1 --> P2[Plan 002: SQLite Storage & Repository]
    P2 --> P3[Plan 003: Task Service Engine]
    P3 --> P4[Plan 004: CLI Interface & Scripting]
    P3 --> P5[Plan 005: Interactive TUI Application]
    P4 --> P6[Plan 006: Automation, Packaging & Release]
    P5 --> P6
```

## Plan Catalog & Ultrathink Planning Packs

| Plan ID | Title | Implementation Plan | Verification Plan | Issue Workorder | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Brainstorm** | Modern Task System | [Brainstorm Plan](2026-09-06-001-feat-tusk-modern-task-system-plan.md) | — | — | `approved` |
| **000** | Setup & Multi-Agent Foundation | [Plan 000](2026-09-06-000-feat-setup-and-multi-agent-foundation-plan.md) | Embedded in Plan | Tested by scripts | `completed` |
| **001** | Core Domain & Invariants | [Plan 001](2026-09-06-001-feat-core-domain-and-invariants-plan.md) | [Verification Plan](../verification-plans/2026-09-06-001-feat-core-domain-and-invariants-verification-plan.md) | [Workorder](../workorders/2026-09-06-001-feat-core-domain-and-invariants-issues-workorder.md) | `ready` |
| **002** | SQLite Storage & Repository | [Plan 002](2026-09-06-002-feat-sqlite-storage-and-repository-plan.md) | Planned | Planned | `pending` |
| **003** | Task Service Engine | [Plan 003](2026-09-06-003-feat-task-service-engine-plan.md) | Planned | Planned | `pending` |
| **004** | CLI Interface & Scripting | [Plan 004](2026-09-06-004-feat-cli-interface-and-scripting-plan.md) | Planned | Planned | `pending` |
| **005** | Interactive TUI Application | [Plan 005](2026-09-06-005-feat-interactive-tui-application-plan.md) | Planned | Planned | `pending` |
| **006** | Automation, Packaging & Release | [Plan 006](2026-09-06-006-feat-automation-packaging-and-release-plan.md) | Planned | Planned | `pending` |
