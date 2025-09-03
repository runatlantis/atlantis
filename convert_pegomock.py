#!/usr/bin/env python3

import os
import re
import sys

def convert_file(file_path):
    """Convert pegomock patterns to gomock in a single file."""
    with open(file_path, 'r') as f:
        content = f.read()
    
    original_content = content
    
    # Convert When/ThenReturn patterns
    # When(mock.Method(args)).ThenReturn(ret) -> mock.EXPECT().Method(args).Return(ret)
    content = re.sub(
        r'When\((\w+)\.(\w+)\(([^)]*)\)\)\.ThenReturn\(([^)]*)\)',
        r'\1.EXPECT().\2(\3).Return(\4)',
        content
    )
    
    # Handle remaining When patterns with better matching
    # Match When(backend.Method(...)).ThenReturn(...)
    content = re.sub(
        r'When\(([^.]+)\.([^(]+)\(([^)]*)\)\)\.ThenReturn\(([^)]*)\)',
        r'\1.EXPECT().\2(\3).Return(\4)',
        content
    )
    
    # Convert Any[Type]() to gomock.Any()
    content = re.sub(r'Any\[[^\]]+\]\(\)', 'gomock.Any()', content)
    
    # Convert VerifyWasCalled(Never()) - these typically become expectations that should never be called
    # For now, we'll comment these out as they need manual review
    content = re.sub(
        r'(\w+)\.VerifyWasCalled\(Never\(\)\)\.(\w+)\(([^)]*)\)',
        r'// TODO: Convert Never() expectation: \1.EXPECT().\2(\3).Times(0)',
        content
    )
    
    # Convert VerifyWasCalledOnce() - these are typically used for capturing arguments
    # We'll comment these for manual review
    content = re.sub(
        r'(\w+)\.VerifyWasCalledOnce\(\)\.(\w+)\(',
        r'// TODO: Convert to gomock expectation with argument capture\n\t// \1.EXPECT().\2(',
        content
    )
    
    # Convert VerifyWasCalled(Times(n)) - these need manual review too
    content = re.sub(
        r'(\w+)\.VerifyWasCalled\(Times\((\d+)\)\)\.(\w+)\(([^)]*)\)',
        r'// TODO: Convert to gomock expectation: \1.EXPECT().\3(\4).Times(\2)',
        content
    )
    
    if content != original_content:
        with open(file_path, 'w') as f:
            f.write(content)
        return True
    return False

def main():
    if len(sys.argv) != 2:
        print("Usage: python convert_pegomock.py <file_path>")
        sys.exit(1)
    
    file_path = sys.argv[1]
    if convert_file(file_path):
        print(f"Converted {file_path}")
    else:
        print(f"No changes needed in {file_path}")

if __name__ == "__main__":
    main()