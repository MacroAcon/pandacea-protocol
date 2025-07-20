#!/usr/bin/env python3
"""
Simple test to check if the Go application can start without getting stuck
"""

import subprocess
import time
import signal
import os
import sys

def test_go_build():
    """Test if the Go application builds successfully"""
    print("🔨 Testing Go build...")
    try:
        result = subprocess.run(
            ["go", "build", "-o", "test_agent", "./cmd/agent"],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            print("✅ Go build successful")
            return True
        else:
            print(f"❌ Go build failed:")
            print(f"STDOUT: {result.stdout}")
            print(f"STDERR: {result.stderr}")
            return False
    except subprocess.TimeoutExpired:
        print("❌ Go build timed out")
        return False
    except FileNotFoundError:
        print("❌ Go not found in PATH")
        return False
    except Exception as e:
        print(f"❌ Go build error: {e}")
        return False

def test_go_run():
    """Test if the Go application can start and exit gracefully"""
    print("\n🚀 Testing Go run (brief startup test)...")
    
    try:
        # Start the application
        process = subprocess.Popen(
            ["go", "run", "./cmd/agent"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        # Wait a few seconds for startup
        time.sleep(3)
        
        # Check if process is still running
        if process.poll() is None:
            print("✅ Application started successfully")
            
            # Send SIGTERM to gracefully shutdown
            process.terminate()
            
            # Wait for graceful shutdown
            try:
                process.wait(timeout=10)
                print("✅ Application shut down gracefully")
                return True
            except subprocess.TimeoutExpired:
                print("⚠️  Application didn't shut down gracefully, forcing...")
                process.kill()
                process.wait()
                return True
        else:
            # Process exited early
            stdout, stderr = process.communicate()
            print(f"❌ Application exited early:")
            print(f"Return code: {process.returncode}")
            print(f"STDOUT: {stdout}")
            print(f"STDERR: {stderr}")
            return False
            
    except FileNotFoundError:
        print("❌ Go not found in PATH")
        return False
    except Exception as e:
        print(f"❌ Go run error: {e}")
        return False

def test_go_test():
    """Test if the tests run successfully"""
    print("\n🧪 Testing Go tests...")
    try:
        result = subprocess.run(
            ["go", "test", "./internal/api", "-v"],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            print("✅ Go tests passed")
            return True
        else:
            print(f"❌ Go tests failed:")
            print(f"STDOUT: {result.stdout}")
            print(f"STDERR: {result.stderr}")
            return False
    except subprocess.TimeoutExpired:
        print("❌ Go tests timed out")
        return False
    except FileNotFoundError:
        print("❌ Go not found in PATH")
        return False
    except Exception as e:
        print(f"❌ Go test error: {e}")
        return False

def main():
    print("=== Pandacea Agent Backend - Startup Test ===\n")
    
    # Test 1: Build
    build_ok = test_go_build()
    
    # Test 2: Tests
    test_ok = test_go_test()
    
    # Test 3: Run (only if build was successful)
    run_ok = False
    if build_ok:
        run_ok = test_go_run()
    
    # Cleanup
    if os.path.exists("test_agent"):
        os.remove("test_agent")
    
    # Summary
    print("\n=== Test Summary ===")
    print(f"Build: {'✅ PASS' if build_ok else '❌ FAIL'}")
    print(f"Tests: {'✅ PASS' if test_ok else '❌ FAIL'}")
    print(f"Run:   {'✅ PASS' if run_ok else '❌ FAIL'}")
    
    if build_ok and test_ok and run_ok:
        print("\n🎉 All tests passed! The application is working correctly.")
    else:
        print("\n⚠️  Some tests failed. There may be issues with the implementation.")
        if not build_ok:
            print("   - Build failure suggests compilation issues")
        if not test_ok:
            print("   - Test failure suggests logic issues")
        if not run_ok:
            print("   - Run failure suggests runtime issues")

if __name__ == "__main__":
    main() 