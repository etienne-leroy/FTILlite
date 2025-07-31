#!/usr/bin/env python3
"""
Script to update the project structure tree.json file.
Scans the current directory and generates a JSON representation
compatible with the existing tree visualization.
"""

import os
import json
import argparse
from pathlib import Path
from datetime import datetime

def should_exclude(name, path):
    """Check if a file or directory should be excluded from the tree."""
    exclude_names = {
        '__pycache__', '.git', '.pytest_cache', 'node_modules', 
        '.vscode', '.idea', 'venv', 'env',
        'postgres-data'  # Keep this exclusion as it has access issues
    }
    
    exclude_patterns = {
        '.pyc', '.pyo', '.pyd', '.dylib', '.dll',
        '.DS_Store', 'Thumbs.db'
    }
    
    # Exclude hidden files/directories (starting with .)
    if name.startswith('.') and name not in {'.', '..'}:
        return True
        
    # Exclude specific names
    if name in exclude_names:
        return True
        
    # Exclude by file extension
    if any(name.endswith(pattern) for pattern in exclude_patterns):
        return True
        
    return False

def scan_directory(path, max_depth=10, current_depth=0):
    """Recursively scan directory and return tree structure."""
    if current_depth > max_depth:
        return {"error": "max depth exceeded"}
        
    try:
        path_obj = Path(path)
        if not path_obj.exists():
            return {"error": "path does not exist"}
            
        if path_obj.is_file():
            return {
                "type": "file",
                "name": path_obj.name
            }
            
        if path_obj.is_dir():
            contents = []
            
            try:
                # Get directory entries and sort them
                entries = sorted(path_obj.iterdir(), 
                               key=lambda x: (x.is_file(), x.name.lower()))
                
                for entry in entries:
                    if should_exclude(entry.name, entry):
                        continue
                        
                    try:
                        if entry.is_dir():
                            subdir_result = scan_directory(entry, max_depth, current_depth + 1)
                            contents.append({
                                "type": "directory",
                                "name": entry.name,
                                "contents": subdir_result.get("contents", [])
                            })
                        elif entry.is_file():
                            contents.append({
                                "type": "file", 
                                "name": entry.name
                            })
                    except PermissionError:
                        contents.append({
                            "type": "directory" if entry.is_dir() else "file",
                            "name": entry.name,
                            "error": "permission denied"
                        })
                    except Exception as e:
                        contents.append({
                            "type": "directory" if entry.is_dir() else "file", 
                            "name": entry.name,
                            "error": f"access error: {str(e)}"
                        })
                        
            except PermissionError:
                return {"error": "opening dir"}
            except Exception as e:
                return {"error": f"scanning error: {str(e)}"}
                
            return {
                "type": "directory",
                "name": path_obj.name,
                "contents": contents
            }
            
    except Exception as e:
        return {"error": f"general error: {str(e)}"}

def count_items(tree_data):
    """Count directories and files in the tree structure."""
    directories = 0
    files = 0
    
    def count_recursive(node):
        nonlocal directories, files
        
        if isinstance(node, dict):
            if node.get("type") == "directory":
                directories += 1
                for item in node.get("contents", []):
                    count_recursive(item)
            elif node.get("type") == "file":
                files += 1
        elif isinstance(node, list):
            for item in node:
                count_recursive(item)
                
    count_recursive(tree_data)
    return directories, files

def main():
    parser = argparse.ArgumentParser(description="Update project structure tree.json")
    parser.add_argument("--path", "-p", default=".", 
                       help="Root path to scan (default: current directory)")
    parser.add_argument("--output", "-o", default="tree.json",
                       help="Output file path (default: tree.json)")
    parser.add_argument("--max-depth", "-d", type=int, default=10,
                       help="Maximum directory depth to scan (default: 10)")
    parser.add_argument("--pretty", action="store_true",
                       help="Pretty print JSON output")
    
    args = parser.parse_args()
    
    print(f"Scanning directory structure from: {os.path.abspath(args.path)}")
    print(f"Maximum depth: {args.max_depth}")
    
    # Scan the directory structure
    root_path = Path(args.path).resolve()
    tree_data = scan_directory(root_path, args.max_depth)
    
    # Count items
    directories, files = count_items(tree_data)
    
    # Create the final structure (as array to match original format)
    result = [
        tree_data,
        {
            "type": "report",
            "directories": directories,
            "files": files,
            "generated": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            "root_path": str(root_path)
        }
    ]
    
    # Write to output file
    output_path = Path(args.output)
    try:
        with output_path.open('w', encoding='utf-8') as f:
            if args.pretty:
                json.dump(result, f, indent=2, ensure_ascii=False)
            else:
                json.dump(result, f, ensure_ascii=False)
                
        print(f"\nTree structure saved to: {output_path.absolute()}")
        print(f"Statistics: {directories} directories, {files} files")
        
    except Exception as e:
        print(f"Error writing output file: {e}")
        return 1
        
    return 0

if __name__ == "__main__":
    exit(main())
