#!/usr/bin/env python3

import os
import re
import subprocess
import sys

def find_test_files():
    """Find all test files that need conversion."""
    result = subprocess.run(['find', '/Users/pepe.amengual/github/atlantis/server', '-name', '*_test.go'], 
                          capture_output=True, text=True)
    return result.stdout.strip().split('\n')

def update_mock_constructors(file_path):
    """Update mock constructors to use gomock controller."""
    with open(file_path, 'r') as f:
        content = f.read()
    
    # Pattern: NewMock*() -> NewMock*(ctrl)
    patterns = [
        (r'(\w+\.)?NewMock\w+\(\)', r'\1NewMock\w+(ctrl)'),
        # More specific patterns
        (r'mocks\.NewMockWorkingDir\(\)', r'mocks.NewMockWorkingDir(ctrl)'),
        (r'mocks\.NewMockLocker\(\)', r'mocks.NewMockLocker(ctrl)'),
        (r'mocks\.NewMockClient\(\)', r'mocks.NewMockClient(ctrl)'),
        (r'mocks\.NewMockBackend\(\)', r'mocks.NewMockBackend(ctrl)'),
        (r'lockmocks\.NewMockLocker\(\)', r'lockmocks.NewMockLocker(ctrl)'),
        (r'vcsmocks\.NewMockClient\(\)', r'vcsmocks.NewMockClient(ctrl)'),
        (r'tfclientmocks\.NewMockClient\(\)', r'tfclientmocks.NewMockClient(ctrl)'),
    ]
    
    modified = False
    for old_pattern, new_pattern in patterns:
        new_content = re.sub(old_pattern, new_pattern, content)
        if new_content != content:
            content = new_content
            modified = True
    
    if modified:
        with open(file_path, 'w') as f:
            f.write(content)
        return True
    return False

def update_when_then_patterns(file_path):
    """Update When/ThenReturn patterns to EXPECT/Return."""
    with open(file_path, 'r') as f:
        content = f.read()
    
    # This is more complex and needs careful handling
    # For now, let's handle simple cases
    patterns = [
        # When(mock.Method(args)).ThenReturn(ret) -> mock.EXPECT().Method(args).Return(ret)
        (r'When\((\w+)\.(\w+)\((.*?)\)\)\.ThenReturn\((.*?)\)', r'\1.EXPECT().\2(\3).Return(\4)'),
        # Any[Type]() -> gomock.Any()
        (r'Any\[\w+\]\(\)', r'gomock.Any()'),
    ]
    
    modified = False
    for old_pattern, new_pattern in patterns:
        new_content = re.sub(old_pattern, new_pattern, content)
        if new_content != content:
            content = new_content
            modified = True
    
    if modified:
        with open(file_path, 'w') as f:
            f.write(content)
        return True
    return False

def main():
    test_files = find_test_files()
    converted_files = []
    
    for file_path in test_files:
        if not file_path:
            continue
            
        print(f"Processing {file_path}...")
        
        # Update mock constructors
        if update_mock_constructors(file_path):
            print(f"  Updated mock constructors in {file_path}")
            
        # Update When/ThenReturn patterns
        if update_when_then_patterns(file_path):
            print(f"  Updated When/ThenReturn patterns in {file_path}")
            
        converted_files.append(file_path)
    
    print(f"\nProcessed {len(converted_files)} files")

if __name__ == "__main__":
    main()