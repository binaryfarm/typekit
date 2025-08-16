import b from 'benny';

function sumArrayForLoop(arr: number[]): number {
	let sum = 0;
	for (let i = 0; i < arr.length; i++) {
		sum += arr[i];
	}
	return sum;
}

function sumArrayReduce(arr: number[]): number {
	return arr.reduce((acc, val) => acc + val, 0);
}

const largeArray = Array.from({ length: 100000 }, (_, i) => i);

b.suite(
	'Array Summation',

	b.add('For Loop', () => {
		sumArrayForLoop(largeArray);
	}),

	b.add('Reduce', () => {
		sumArrayReduce(largeArray);
	}),

	b.cycle(), // Runs after each test in the suite
	b.complete(), // Runs after all tests in the suite are done
	b.save({ file: 'array-sum-benchmark', format: 'json' }) // Optional: saves results
);