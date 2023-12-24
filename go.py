import os
import subprocess
import threading

def run_go_program():
    try:
        subprocess.run(["go", "run", "."], check=True)
    except subprocess.CalledProcessError as e:
        print(f"Error running Go program: {e}")

def main():
    go_thread = threading.Thread(target=run_go_program)
    go_thread.start()
    # go_thread.join()

if __name__ == "__main__":
    main()