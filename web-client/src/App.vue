<template>
  <v-app>
    <v-tabs centered>
      <v-tab to="SyncView">Sync</v-tab>
      <v-tab>Buyers</v-tab>
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
