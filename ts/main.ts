function serialiseString(s: string): Uint8Array {
	const bs = Buffer.from(s, "utf8")
	const res = Buffer.alloc(4 + bs.length)
	res.writeUInt32BE(bs.length, 0)
	bs.copy(res, 4)
	return res
}

export class BufReader {
	off = 0
	constructor(private buf: Buffer) {}
	readExact(n: number): Buffer {
		if (this.buf.length - this.off < n)
			throw new Error("not enough bytes to read")

		const r = this.buf.subarray(this.off, this.off + n)
		this.off += n
		return r
	}
}

export const ErrDataTooShort = new Error("data too short to decode")
export const ErrDataTooShortLength = new Error(
	`${ErrDataTooShort.message} length`
)
export const ErrDataTooShortString = new Error(
	`${ErrDataTooShort.message} string`
)

function getLen(r: BufReader): number {
	try {
		return r.readExact(4).readUInt32BE(0)
	} catch {
		throw ErrDataTooShortLength
	}
}

function getStr(r: BufReader, l: number): string {
	try {
		return r.readExact(l).toString("utf8")
	} catch {
		throw ErrDataTooShortString
	}
}

function deserialiseString(r: BufReader): string {
	return getStr(r, getLen(r))
}

export type ToSub = { [_: string]: string }

export class PartName {
	constructor(
		public Part: string,
		public Name: string
	) {}

	serialise(): Uint8Array {
		return Buffer.concat([
			serialiseString(this.Part),
			serialiseString(this.Name),
		])
	}

	equals(other: PartName): boolean {
		return this.Part === other.Part && this.Name === other.Name
	}

	static deserialise(r: BufReader): PartName {
		return new PartName(deserialiseString(r), deserialiseString(r))
	}
}

export const ErrMissingValue = new Error("missing value for name")
export const ErrExtraValue = new Error("extra value provided for name")

export class Sub {
	constructor(
		public PartNames: PartName[],
		public Final: string
	) {}

	Sub(to: ToSub): string {
		const seen = new Set<string>()
		for (const pn of this.PartNames) {
			const n = pn.Name
			if (to[n] === undefined)
				throw new Error(`${ErrMissingValue.message}: ${n}`)
			seen.add(n)
		}

		for (const k in to) {
			if (!seen.has(k)) throw new Error(`${ErrExtraValue.message}: ${k}`)
		}

		const b: string[] = []

		for (const pn of this.PartNames) {
			const val = to[pn.Name]
			if (val === undefined)
				// go moment bro moment
				throw new Error("Missing value should have already been caught")

			b.push(pn.Part)
			b.push(val)
		}

		b.push(this.Final)

		return b.join("")
	}

	Equals(other: Sub): boolean {
		if (this.PartNames.length !== other.PartNames.length) return false

		for (let i = 0; i < this.PartNames.length; i++) {
			const p1 = this.PartNames[i]
			const p2 = other.PartNames[i]
			if (!p1 || !p2 || !p1.equals(p2)) return false
		}

		return this.Final === other.Final
	}

	Serialise(): Uint8Array {
		const countBuf = Buffer.alloc(4)
		countBuf.writeUInt32BE(this.PartNames.length, 0)
		const parts: Uint8Array[] = [countBuf]
		for (const pn of this.PartNames) {
			parts.push(pn.serialise())
		}
		parts.push(serialiseString(this.Final))
		return Buffer.concat(parts)
	}

	static Deserialise(input: Uint8Array | Buffer): Sub {
		const r = new BufReader(Buffer.from(input))
		const l = getLen(r)

		const partNames: PartName[] = []
		for (let i = 0; i < l; i++) partNames.push(PartName.deserialise(r))

		return new Sub(partNames, deserialiseString(r))
	}
}
