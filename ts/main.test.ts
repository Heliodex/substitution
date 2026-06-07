import { expect, test } from "bun:test"
import * as sub from "./main"

// Test multiple substitutions of the same name
const s = new sub.Sub(
	[
		new sub.PartName("Hello ", "name"),
		new sub.PartName("! You have ", "count"),
		new sub.PartName(" new messages. Thanks, ", "name"),
	],
	"!"
)

test("Substitute", () => {
	const toSub = { name: "Heliodex", count: "67" } as sub.ToSub
	const result = s.Sub({ ...toSub })
	expect(result).toBe("Hello Heliodex! You have 67 new messages. Thanks, Heliodex!")
})

test("SubstituteMissing", () => {
	const toSub = { name: "Heliodex" } as sub.ToSub
	let err: Error | null = null
	try {
		s.Sub({ ...toSub })
	} catch (e) {
		err = e as Error
	}
	expect(err).not.toBeNull()
	expect(err?.message).toContain(sub.ErrMissingValue.message)
})

test("SubstituteExtra", () => {
	const toSub = { name: "Heliodex", count: "67", extra: "value" } as sub.ToSub
	let err: Error | null = null
	try {
		s.Sub({ ...toSub })
	} catch (e) {
		err = e as Error
	}
	expect(err).not.toBeNull()
	expect(err?.message).toContain(sub.ErrExtraValue.message)
})

test("Serialisation", () => {
	const data = s.Serialise()
	const s2 = sub.Sub.Deserialise(Buffer.from(data))
	expect(s.Equals(s2)).toBe(true)
})

test("DeserialisationTooShortLength", () => {
	let err: Error | null = null
	try {
		sub.Sub.Deserialise(Buffer.from([0, 0, 0]))
	} catch (e) {
		err = e as Error
	}
	expect(err).not.toBeNull()
	expect(err).toBe(sub.ErrDataTooShortLength)
})

test("DeserialisationTooShortString", () => {
	let err: Error | null = null
	try {
		sub.Sub.Deserialise(Buffer.from([0, 0, 0, 1, 0, 0, 0, 1]))
	} catch (e) {
		err = e as Error
	}
	expect(err).not.toBeNull()
	expect(err).toBe(sub.ErrDataTooShortString)
})

test("Equals", () => {
	const s2 = new sub.Sub(
		[
			new sub.PartName("Hello ", "name"),
			new sub.PartName("! You have ", "count"),
			new sub.PartName(" new messages. Thanks, ", "name"),
		],
		"!"
	)

	expect(s.Equals(s2)).toBe(true)

	const s3 = new sub.Sub(
		[
			new sub.PartName("Sup ", "name"),
			new sub.PartName(", you got ", "num"),
			new sub.PartName(" new pings. Thanks, ", "name"),
		],
		"!"
	)

	expect(s.Equals(s3)).toBe(false)
})
