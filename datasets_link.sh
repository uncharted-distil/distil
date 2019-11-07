#!/bin/bash

# datasets we want
declare -a datasets_seed=(32_wikiqa 185_baseball 196_autoMpg 66_chlorineConcentration 22_handgeometry 56_sunspots LL1_336_MS_Geolife_transport_mode_prediction_separate_lat_lon LL0_acled_reduced_clean world_bank_2018 LL1_736_population_spawn SEMI_1040_sylva_prior SEMI_1044_eye_movements SEMI_1053_jm1 SEMI_1217_click_prediction_small SEMI_1459_artificial_characters SEMI_155_pokerhand 124_214_coil20 state_immigration_representation incarceration LL1_terra_canopy_height_long_form_s4_90)
declare -a datasets_aug=(DA_college_debt DA_ny_taxi_demand DA_medical_malpractice DA_poverty_estimation)
declare -a datasets_eval=(LL0_USER_EVAL_TASK1_1100_popularkids)

# source locations of the datasets
seed_location="/data/datasets/seed_datasets_current"
aug_location="/data/datasets/seed_datasets_data_augmentation"
eval_location="/data/datasets/seed_datasets_user_eval"

# create the target folder
target_location="/data/merged/"
mkdir $target_location

# link seed datasets
for dataset in "${datasets_seed[@]}"
do
   source=$seed_location/$dataset
   target=$target_location/$dataset
   echo "making link from $source to $target"

   ln -s source target
done

# link aug datasets
for dataset in "${datasets_aug[@]}"
do
   source=$aug_location/$dataset
   target=$target_location/$dataset
   echo "making link from $source to $target"

   ln -s source target
done

# link eval datasets
for dataset in "${datasets_eval[@]}"
do
   source=$eval_location/$dataset
   target=$target_location/$dataset
   echo "making link from $source to $target"

   ln -s source target
done
