'use strict';

const units = ['y', 'z', 'a', 'f', 'p', 'n', 'µ', 'm', '', 'k', 'M', 'G', 'T', 'P', 'E', 'Z', 'Y']

const Metric = new class {
	format(v, prefix) {
		if (v == 0) { return v.toFixed(3).concat(prefix) }

		let y = Math.trunc(Math.log10(Math.abs(v)) / 3)
		if (y < -8) { y = -8 } else if (y > 8) { y = 8 }
		v *= Math.pow(1000, -y)

		// calculate number of digits right of decimal place
		// x < 0 for minus sign
		// z for digits left of decimal place
		// z < 4  for decimal point
		// y != 0 for metric prefix
		const z = Math.trunc(Math.log10(Math.abs(v)))
		let p = 6 - (v < 0) - z - (z < 4) - (y != 0) - prefix.length
		if (p < 0) { p = 0 } else if (p > 3) { p = 3 } 
		return v.toFixed(p).concat(units[y + 8], prefix)
	}
}

class formatter { 
	constructor(prefix) { this.prefix = prefix }
	format(s, x) { return s.format(x, this.prefix) }
}

const Voltage = new formatter('V')
const Current = new formatter('A')
const Resistance = new formatter('Ω')
const Capacitance = new formatter('F')
const Frequency = new formatter('Hz') // '㎐'
const Power = new formatter('W')
