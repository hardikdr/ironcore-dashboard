import { describe, expect, it } from 'vitest'
import { classifyStatus, statusPresentation } from './status'

describe('status semantics', () => {
  it.each([['Running','healthy'],['Degraded','warning'],['FAILED','error'],['Provisioning','pending'],[undefined,'unknown'],['','unknown'],['Surprising','unknown']])('classifies %s conservatively', (value, expected) => expect(classifyStatus(value)).toBe(expected))
  it('keeps an unrecognized raw state visible', () => expect(statusPresentation('Surprising')).toMatchObject({ label:'Surprising', severity:'unknown' }))
})
