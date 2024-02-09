import os

def print_project_structure(root_path='.'):
    structure = {}
    root_path = os.path.abspath(root_path)  # Ensure absolute path

    for root, dirs, files in os.walk(root_path, topdown=True):
        dirs[:] = [d for d in dirs if d != '.git']  # Skip .git directory
        # Trim the root_path part and ensure a relative path representation
        parts = os.path.relpath(root, root_path).split(os.path.sep) if os.path.relpath(root, root_path) != '.' else []
        current = structure
        for part in parts:
            current = current.setdefault(part, {})
        for file in files:
            if '.git' not in root:  # Additionally check to avoid adding files from within .git
                current[file] = None

    def print_dict(d, indent=0):
        for key, value in d.items():
            print('  ' * indent + ('/' if indent else '') + key)
            if isinstance(value, dict) and value:
                print_dict(value, indent + 1)
            elif isinstance(value, dict) and not value and indent:  # Empty directory
                print('  ' * (indent + 1) + "(empty)")

    print_dict(structure)

print_project_structure()
