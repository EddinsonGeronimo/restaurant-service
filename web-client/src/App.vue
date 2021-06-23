<template>
  <v-app>
    <v-tabs centered>
      <v-tab to="/">Sync</v-tab>
      <v-tab to="/buyers">Buyers</v-tab>
    </v-tabs>
    <v-main>
      <router-view/>
    </v-main>
  </v-app>
</template>

<script>

export default {
  name: 'App',

  data: () => ({
    //
  }),
};
</script>
