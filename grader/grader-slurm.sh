#!/bin/bash
#
#SBATCH --mail-user=ywang27@cs.uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj1_grade 
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --chdir=/home/ywang27/parallel/proj1-wywy136/proj1/grader
#SBATCH --partition=debug
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=10:00

module load golang/1.16.2 
go run proj1/grader proj1
