import { describe, expect, it } from 'vitest'
import { shallowMount } from '@vue/test-utils'

import { SideBar } from '@/components'

describe('SideBar.vue', () => {
  it('should render', () => {
    const wrapper = shallowMount(SideBar)

    expect(wrapper).toBeDefined()
  })
})
