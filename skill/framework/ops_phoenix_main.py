#!/usr/bin/env python3
"""
Ops Phoenix - Entry Point

This is the main script users should run. It handles:
1. Loading configuration
2. Asking setup questions if needed
3. Running the ops agent

Usage:
    python3 ops_phoenix.py              # Interactive (asks if incomplete)
    python3 ops_phoenix.py --setup     # Force setup wizard
    python3 ops_phoenix.py --dry-run   # Detect errors only
    python3 ops_phoenix.py --full      # Full cycle
"""

import os
import sys
from pathlib import Path

# Change to framework directory
FRAMEWORK_DIR = Path(__file__).parent
os.chdir(FRAMEWORK_DIR)

# Run the main agent
sys.path.insert(0, str(FRAMEWORK_DIR))
from ops_phoenix import main

if __name__ == "__main__":
    main()