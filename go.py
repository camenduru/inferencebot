import os
import subprocess
import threading

def set_environment_variables():
    os.environ["PATH"] = os.environ.get("PATH", "") + ":/content/go/bin"
    os.environ["GOPATH"] = os.path.expanduser("~/go")
    os.environ["PATH"] = os.environ.get("PATH", "") + ":" + os.path.join(os.environ["GOPATH"], "bin")

def run_go_program():
    try:
        subprocess.run(["go", "run", "."], check=True)
    except subprocess.CalledProcessError as e:
        print(f"Error running Go program: {e}")

def main():
    set_environment_variables()
    go_thread = threading.Thread(target=run_go_program)
    go_thread.start()
    # go_thread.join()

if __name__ == "__main__":
    main()