# Project Values:

- a functional style that values immutability is preferred in all but absolutely necessary scenarios (use of pointers and mutable methods should be rare and well-justified)
- an emphasis on "deep interfaces" - that is, avoidance of unnecessary interfaces wrapping little complexity underneath. Interfaces should have as few public methods with as low an arity as possible. Choosing the right interfaces at the right level of abstraction is crucial for this.
- there is a strong emphasis on unit tests with a consistent style. Tests should only assert on publicly exposed fields and results of public methods. Fields or methods are never be made public only for the purpose of testing.
- once a feature is complete, code comments should be reviewed. Final code comments should focus on the "why", not the "what"
- heavily relies on https://github.com/robinovitch61/bubbleo/tree/main/filterableviewport
- all styling and layout is done with https://github.com/charmbracelet/lipgloss

# Project Overview:

See the README.md for a public-facing overview of what this project is.

See the SPECIFICATION.md ("the spec") for the master reference for how this project should behave. If code being implemented is not in the spec or conflicts with the spec, clarify this discrepancy with the user immediately. All new feature work and changes should be included in the spec before code is (re)written.

See INTERFACES.md for a description of the most important interfaces (packages, structs, GoLang interfaces, key methods, etc.). 