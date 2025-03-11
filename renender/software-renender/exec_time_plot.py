import numpy as np
import matplotlib.pyplot as plt

def read_times(filename):
    with open(filename, 'r') as f:
        times = [int(line.strip()) for line in f]
    return np.array(times)

def plot_times(times_one, times_two):
    plt.figure(figsize=(10, 5))

    plt.plot(times_one, label='multi-threading')
    plt.plot(times_two, label='single-threading')
    plt.title('Function Execution Times Comparison')
    plt.xlabel('Iteration')
    plt.ylabel('Time (microseconds)')
    plt.legend()

    plt.tight_layout()
    plt.show()

times_one = np.random.normal(loc=50, scale=10, size=100)
times_two = np.random.normal(loc=100, scale=20, size=100)



times_one = read_times('times_multi-threading.txt')
times_two = read_times('times_single-threading.txt')

plot_times(times_one, times_two)