import subprocess
import matplotlib.pyplot as plt
import time


settings = [
    "xsmall",
    "small",
    "medium",
    "large",
    "xlarge"
]
x = [0, 1, 2, 3, 4]
threadNums = ["2", "4", "6", "8", "12"]

# Sequential version
seqentialTime = dict()
for setting in settings:
    print(f"Now testing sequential - {setting}", flush=True)
    sumTime = 0
    for i in range(5):
        pipe = subprocess.Popen(
            ["go", "run", "benchmark.go", "s", setting],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        stdout, stderr = pipe.communicate()
        sumTime += float(stdout.decode())
    seqentialTime[setting] = sumTime / 5

# Parallel version
parallelTime = dict()
for setting in settings:
    speedUp = []
    for threadNum in threadNums:
        print(f"Now testing parallel - {setting} - {threadNum}", flush=True)
        sumTime = 0
        for i in range(5):
            pipe = subprocess.Popen(
                ["go", "run", "benchmark.go", "p", setting, threadNum],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE
            )
            stdout, stderr = pipe.communicate()
            sumTime += float(stdout.decode())
        sup = seqentialTime[setting] / (sumTime / 5)
        speedUp.append(sup)
    parallelTime[setting] = speedUp
    
fig, ax = plt.subplots()
xsmall, = ax.plot(x, parallelTime["xsmall"], linestyle='--', marker='o')
xsmall.set_label("xsmall")

small, = ax.plot(x, parallelTime["small"], linestyle='--', marker='o')
small.set_label("small")

medium, = ax.plot(x, parallelTime["medium"], linestyle='--', marker='o')
medium.set_label("medium")

large, = ax.plot(x, parallelTime["large"], linestyle='--', marker='o')
large.set_label("large")

xlarge, = ax.plot(x, parallelTime["xlarge"], linestyle='--', marker='o')
xlarge.set_label("xlarge")

ax.legend(loc='upper left')
ax.set_xticks(x)
ax.set_xticklabels(threadNums)
plt.xlabel('Number of Threads')
plt.ylabel('Speedup')
plt.savefig('plot1.png')