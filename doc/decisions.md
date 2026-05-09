# Decisions

## Why record decisions

Write down key development decisions while the context is fresh. A short note today can save hours later by explaining what was chosen, what was rejected, and why the trade-off made sense at the time.

## Guidance

Use a lightweight Architecture Decision Record (ADR) style:

* Record decisions that affect architecture, data flow, public APIs, dependencies, deployment, security, or long-term maintenance.
* Write the decision when it is made, not after the context has faded.
* Prefer short entries that explain the context, decision, alternatives, and consequences.
* Include enough reasoning for a future maintainer to understand the trade-off.
* Do not document every small implementation detail; focus on choices that would be costly or confusing to rediscover.
* Update or supersede earlier decisions instead of silently rewriting history.

## Entry template

```markdown
### YYYY-MM-DD: Decision title

**Status:** Proposed | Accepted | Superseded

**Context:** What problem, constraint, or trade-off led to this decision?

**Decision:** What was chosen?

**Alternatives considered:** What other options were rejected, and why?

**Consequences:** What becomes easier, harder, riskier, or more constrained?
```

## Actual decisions

### 2026-05-03: Target Raspberry Pi Zero 2 WH with Raspberry Pi OS Lite 64-bit

**Status:** Accepted

**Context:** The project needs a minimal, reliable operating system target for deploying the Go bot on a Raspberry Pi Zero 2 WH. The main requirement is to run a compiled Go application with low operational overhead. The main options considered were Raspberry Pi OS Lite, DietPi, Alpine Linux, and Arch Linux ARM. The original Raspberry Pi Zero constraints around ARMv6 do not apply to the Zero 2 WH, which uses an ARMv8 CPU.

**Decision:** Standardize on Raspberry Pi OS Lite 64-bit as the deployment target for the Raspberry Pi Zero 2 WH, and build the bot as a `linux/arm64` binary.

**Alternatives considered:** Raspberry Pi OS Lite 32-bit would be a more conservative compatibility choice, but it adds no clear benefit for a pure Go bot on ARMv8 hardware. DietPi is attractive for aggressive minimalism, but it adds project-specific tooling on top of Debian without solving a real problem for this deployment. Alpine Linux is smaller, but its `musl`-based environment introduces avoidable compatibility risk if the project later adds native dependencies. Arch Linux ARM is flexible, but it is a weaker default for a small deployment-focused project that benefits more from predictability than from distribution minimalism for its own sake.

**Consequences:** This keeps the runtime small while preserving the official Raspberry Pi kernel, package ecosystem, and hardware support path. It also simplifies deployment by aligning the project with a standard `arm64` Linux target. The trade-off is that the base image is not as stripped down as Alpine or DietPi, but the operational risk is lower and the environment is more conventional.
