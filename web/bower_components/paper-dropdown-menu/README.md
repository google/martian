
<!---

This README is automatically generated from the comments in these files:
paper-dropdown-menu-light.html  paper-dropdown-menu.html

Edit those files, and our readme bot will duplicate them over here!
Edit this file, and the bot will squash your changes :)

The bot does some handling of markdown. Please file a bug if it does the wrong
thing! https://github.com/PolymerLabs/tedium/issues

-->

[![Build status](https://travis-ci.org/PolymerElements/paper-dropdown-menu.svg?branch=master)](https://travis-ci.org/PolymerElements/paper-dropdown-menu)

_[Demo and API docs](https://elements.polymer-project.org/elements/paper-dropdown-menu)_


##&lt;paper-dropdown-menu&gt;

Material design: [Dropdown menus](https://www.google.com/design/spec/components/buttons.html#buttons-dropdown-buttons)

`paper-dropdown-menu` is similar to a native browser select element.
`paper-dropdown-menu` works with selectable content. The currently selected
item is displayed in the control. If no item is selected, the `label` is
displayed instead.

Example:

```html
<paper-dropdown-menu label="Your favourite pastry">
  <paper-listbox class="dropdown-content">
    <paper-item>Croissant</paper-item>
    <paper-item>Donut</paper-item>
    <paper-item>Financier</paper-item>
    <paper-item>Madeleine</paper-item>
  </paper-listbox>
</paper-dropdown-menu>
```

This example renders a dropdown menu with 4 options.

The child element with the class `dropdown-content` is used as the dropdown
menu. This can be a [`paper-listbox`](paper-listbox), or any other or
element that acts like an [`iron-selector`](iron-selector).

Specifically, the menu child must fire an
[`iron-select`](iron-selector#event-iron-select) event when one of its
children is selected, and an [`iron-deselect`](iron-selector#event-iron-deselect)
event when a child is deselected. The selected or deselected item must
be passed as the event's `detail.item` property.

Applications can listen for the `iron-select` and `iron-deselect` events
to react when options are selected and deselected.

### Styling

The following custom properties and mixins are also available for styling:

| Custom property | Description | Default |
| --- | --- | --- |
| `--paper-dropdown-menu` | A mixin that is applied to the element host | `{}` |
| `--paper-dropdown-menu-disabled` | A mixin that is applied to the element host when disabled | `{}` |
| `--paper-dropdown-menu-ripple` | A mixin that is applied to the internal ripple | `{}` |
| `--paper-dropdown-menu-button` | A mixin that is applied to the internal menu button | `{}` |
| `--paper-dropdown-menu-input` | A mixin that is applied to the internal paper input | `{}` |
| `--paper-dropdown-menu-icon` | A mixin that is applied to the internal icon | `{}` |

You can also use any of the `paper-input-container` and `paper-menu-button`
style mixins and custom properties to style the internal input and menu button
respectively.



##&lt;paper-dropdown-menu-light&gt;

Material design: [Dropdown menus](https://www.google.com/design/spec/components/buttons.html#buttons-dropdown-buttons)

This is a faster, lighter version of `paper-dropdown-menu`, that does not
use a `<paper-input>` internally. Use this element if you're concerned about
the performance of this element, i.e., if you plan on using many dropdowns on
the same page. Note that this element has a slightly different styling API
than `paper-dropdown-menu`.

`paper-dropdown-menu-light` is similar to a native browser select element.
`paper-dropdown-menu-light` works with selectable content. The currently selected
item is displayed in the control. If no item is selected, the `label` is
displayed instead.

Example:

```html
<paper-dropdown-menu-light label="Your favourite pastry">
  <paper-listbox class="dropdown-content">
    <paper-item>Croissant</paper-item>
    <paper-item>Donut</paper-item>
    <paper-item>Financier</paper-item>
    <paper-item>Madeleine</paper-item>
  </paper-listbox>
</paper-dropdown-menu-light>
```

This example renders a dropdown menu with 4 options.

The child element with the class `dropdown-content` is used as the dropdown
menu. This can be a [`paper-listbox`](paper-listbox), or any other or
element that acts like an [`iron-selector`](iron-selector).

Specifically, the menu child must fire an
[`iron-select`](iron-selector#event-iron-select) event when one of its
children is selected, and an [`iron-deselect`](iron-selector#event-iron-deselect)
event when a child is deselected. The selected or deselected item must
be passed as the event's `detail.item` property.

Applications can listen for the `iron-select` and `iron-deselect` events
to react when options are selected and deselected.

### Styling

The following custom properties and mixins are also available for styling:

| Custom property | Description | Default |
| --- | --- | --- |
| `--paper-dropdown-menu` | A mixin that is applied to the element host | `{}` |
| `--paper-dropdown-menu-disabled` | A mixin that is applied to the element host when disabled | `{}` |
| `--paper-dropdown-menu-ripple` | A mixin that is applied to the internal ripple | `{}` |
| `--paper-dropdown-menu-button` | A mixin that is applied to the internal menu button | `{}` |
| `--paper-dropdown-menu-icon` | A mixin that is applied to the internal icon | `{}` |
| `--paper-dropdown-menu-disabled-opacity` | The opacity of the dropdown when disabled | `0.33` |
| `--paper-dropdown-menu-color` | The color of the input/label/underline when the dropdown is unfocused | `--primary-text-color` |
| `--paper-dropdown-menu-focus-color` | The color of the label/underline when the dropdown is focused | `--primary-color` |
| `--paper-dropdown-error-color` | The color of the label/underline when the dropdown is invalid | `--error-color` |
| `--paper-dropdown-menu-label` | Mixin applied to the label | `{}` |
| `--paper-dropdown-menu-input` | Mixin appled to the input | `{}` |

Note that in this element, the underline is just the bottom border of the "input".
To style it:

```html
<style is=custom-style>
  paper-dropdown-menu-light.custom {
    --paper-dropdown-menu-input: {
      border-bottom: 2px dashed lavender;
    };
</style>
```


