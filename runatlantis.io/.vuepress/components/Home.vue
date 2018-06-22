<template>
  <div class="home">
    <div class="hero">
      <img :src="$withBase(data.heroImage)" alt="hero">
      <h1>{{ data.heroText || $title || 'Hello' }}</h1>
      <p class="description">
        {{ data.tagline || $description || 'Welcome to your VuePress site' }}
      </p>
      <p class="action" >
        <NavLink class="action-button" :item="actionLink"/>
      </p>
    </div>
    <div class="features">
      <div class="feature">
        <h2>{{ feature.title }}</h2>
        <p>{{ feature.details }}</p>
      </div>
    </div>
    <Content custom/>
  </div>
</template>
