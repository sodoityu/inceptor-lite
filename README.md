## Inceptor Lite Architecture

graph TD
    A[main.go] --> B[omc: Must-Gather & Operator Analyzer]
    A --> C[kcs: KCS Knowledge Search]
    A --> D[ai: AI Analysis & Suggestions]
    A --> E[ui: HTML & File Output]
    A --> F[utils: Common Utilities]

    B --> B1[Collect Must-Gather Data]
    B --> B2[Check Cluster Operators]
    B --> B3[Collect Pod & Node Logs]

    C --> C1[KCS Search by Keywords]
    C --> C2[Integrate Search Results]

    D --> D1[Log-based AI Analysis]
    D --> D2[Operator-based AI Analysis]

    E --> E1[Generate HTML Report]
    E --> E2[Save to Local File]

    F --> F1[Logging Helpers]
    F --> F2[Shell Command Wrappers]












