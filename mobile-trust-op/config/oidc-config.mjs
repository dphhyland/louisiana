const clients = [
  {
    client_id: "localhost",
    client_secret: "test-secret",
    grant_types: ['client_credentials'],
    redirect_uris: [], // This client does not have a redirect URI as it's for client_credentials flow
  },
];

export default {
  clients,
  formats: {
    AccessToken: 'jwt', // Use JWT format for access tokens
  },
  jwks: {
    keys: [
      {
        "kty": "RSA",
        "kid": "O6Sd6RUD1T1HlduakgjDFVgz_70-2S1vQAQrn2huaxg",
        "n": "rdwZNKjKpKfKjSAVg_g-NBTr_vGXaAgp5QEggavYIBlAJrtAyYe_KLKKIOYZ9rckW2ztc7aWODfg2wX0ahozav8OdVOVDFtTETxSwTFTADjYTtuEeTbnBDwQhdq1u6QlI7XHku5eUrYS0Xj8qTe9zeWEoRgGl0Vsy0ns396qBUJBTcg-rXudZ3TxwJJr6OUyyQCucYx3Mh6oSocsuAyJuRKWyscH2i4G6IcJN1plEFkerR_Qdfi23ATn8WBwT-NHXHjAY18KXnZJmNNRmtg2CC6GRycuuqvGVYsKHdjG9lDYz0eR5GfUneVfXFU3wjFdUwwbLpweP7f9jvhmLMRWpw",
        "e": "AQAB",
        "d": "VEPdI5y7utpoXcYzJ2dmHrAVQdeuXom2ZHwLWU4EzmnuodcYK5VTnVILiK593woc4QMGg5L3dAABR6a3M8XHLhC43TzJfNe2hyAJrAFQDUd_75iNuIJXrcG-8GF0u6FvLsOoGxrVNJlyvXw6WXne0LBxe_K9HwxTEqSd3lN5bFpH4oh2X8tW251UfqQV2sCsSSx6iYID-RdPpinH2vhDM5-I9S5futXV5jOou9HkJObi5cEwcClaE-mIok8CMv8RORzM4jyG3Z9Fd1sEPrYKcTpLnQMCOi7KY9oY2tAMq5KjkYuK43xgWhdcBvFPekzqwOi-8KaKwp_AlJceUrP3wQ",
        "p": "2Fi3EyzT_az8Wu3EkdfyLrudN1nO4Pcz1STnFLSw90tCqt2KUDLo_2nI1oRKnKTYNJix1q5M7xiFR5H_kt6b5Bzcrg_jgBYxFWxv3xyFugUfn6iSgjh75L6-ifZgVMQIsEAVJaJHP3NAg4efCdPnOwcUj-5QyexqarsfLcxUoc8",
        "q": "zbnX8MuogVYtie23zBWad2qU65V-31pCFiAJZfU8kvIYCGMKEIXsVlksZTAfe6OrZ8UG_Zr5nc_px-34BhQBPI_NBk8mvvlFRo9avwtPzKT7hOmTrj1iKuBciH8V3gCkLAHfsoMf3_AzApgoihoIdQlNpvbgf4v9LdLqgid8a6k",
        "dp": "b5GIj-3tQQPle-rkFSYc8ba1A-dfhapV45RNSuNH46-_KKho_KTUfWsLNH0ykQd9q9oW1BQ8_oxRpzAGcbI1CHVN5MGy28oV8-tg4dkuqVidD1P6ggco9TTcw_73fJ8_r8zMWwUN4w3Hdk0JWiMaOtmS5ArNROSTmIj5MkAOYQc",
        "dq": "tAHo8MlCp-i_7_AQ4oIEpnib1Yb5fHV9Xd6qWow6dFlgrfX62HpWUTe1oNR6t7R-Rk4fz8fKoktKQ6X1X1whuNwaiWq4fGGgPB0zSfab3HR_n8zBa7RKMTofvU910sq828TH92rEeS2zdJGu5yfgPoflajKaPQBAh9gFWd0zAQ",
        "qi": "hf4pvICwi-3f7NlWg0YswkG4RFCQ2k6mWrhdtd3R1VTThL2oft4Z4aEfcB3IqR7ikmLqXHErByxAdSpz8VOW4TIItcNo--FzdEszJPthdKb7s3i6qLZHq6Y9sDCBkmZpzrRqSFP4VUCngLgjt_NItlWMjdVdP0_LKWs49eE91fc"
      },
    ],
  },
  features: {
    fapi: { enabled: true, profile: '2.0' },  // Enables FAPI for financial-grade API security
  },
  cookies: {
    keys: ['your-secure-cookie-key1', 'your-secure-cookie-key2'], // Secure keys for signing cookies
  },
  proxy: true, 
};