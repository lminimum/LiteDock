import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import AuthCard from '@/components/auth/AuthCard.vue'

describe('AuthCard', () => {
  it('renders header slot content', () => {
    const wrapper = mount(AuthCard, {
      slots: {
        header: '<h1>Welcome Back</h1>',
      },
    })

    const header = wrapper.find('.auth-header')
    expect(header.exists()).toBe(true)
    expect(header.text()).toContain('Welcome Back')
  })

  it('renders default slot content', () => {
    const wrapper = mount(AuthCard, {
      slots: {
        default: '<p>Sign in to your account</p>',
      },
    })

    const content = wrapper.find('.auth-content')
    expect(content.exists()).toBe(true)
    expect(content.text()).toContain('Sign in to your account')
  })

  it('renders footer slot content', () => {
    const wrapper = mount(AuthCard, {
      slots: {
        footer: '<span>Need help?</span>',
      },
    })

    const footer = wrapper.find('.auth-footer')
    expect(footer.exists()).toBe(true)
    expect(footer.text()).toContain('Need help?')
  })

  it('renders all three slots simultaneously', () => {
    const wrapper = mount(AuthCard, {
      slots: {
        header: '<h1>Login</h1>',
        default: '<form>email + password fields</form>',
        footer: '<a href="/register">Create account</a>',
      },
    })

    const header = wrapper.find('.auth-header')
    const content = wrapper.find('.auth-content')
    const footer = wrapper.find('.auth-footer')

    expect(header.exists()).toBe(true)
    expect(header.text()).toContain('Login')

    expect(content.exists()).toBe(true)
    expect(content.text()).toContain('email + password fields')

    expect(footer.exists()).toBe(true)
    expect(footer.text()).toContain('Create account')
  })

  it('does not render header div when no header slot is provided', () => {
    const wrapper = mount(AuthCard, {
      slots: {
        default: '<p>Main content only</p>',
      },
    })

    expect(wrapper.find('.auth-header').exists()).toBe(false)
    expect(wrapper.find('.auth-content').exists()).toBe(true)
  })

  it('does not render footer div when no footer slot is provided', () => {
    const wrapper = mount(AuthCard, {
      slots: {
        default: '<p>Main content only</p>',
      },
    })

    expect(wrapper.find('.auth-footer').exists()).toBe(false)
    expect(wrapper.find('.auth-content').exists()).toBe(true)
  })
})
